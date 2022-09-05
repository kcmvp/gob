package builder

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
)

const (
	testCoverOut     = "cover.out"
	testCoverReport  = "cover.html"
	testPackageCover = "cover.json"
)

//go:embed template/*.tmpl
var templateDir embed.FS

var genBuilder boot.Action = func(project *boot.Project, cmd string) error {
	log.Println("Creating project build file")
	var err error
	var tf []byte
	tf, err = templateDir.ReadFile(filepath.Join("template", "builder.tmpl"))
	if err == nil {
		err = boot.GenerateFile(string(tf), filepath.Join(project.ScriptDir(), "builder.go"), nil, false)
	}
	if err != nil {
		err = fmt.Errorf("failed to generate builder script:%w", err)
	}
	return err
}

var getHook boot.Action = func(project *boot.Project, cmd string) error {
	err := genGitHooks(project.GitHome(), project.ScriptDir())
	if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if cmd == "gitHook" {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var linter boot.Action = func(project *boot.Project, action string) error {
	linter := newLinter()
	version := project.GetString("version")
	if ver, err := linter.Configured(project); err != nil {
		if v, err := linter.Install(version); err == nil {
			genLinterCfg(v, false)
		}
	} else {
		linter.Install(ver)
	}
	return nil
}

var createDirAction boot.Action = func(project *boot.Project, cmd string) error {
	var dir string
	// todo fix the bug
	switch cmd {
	case "gitHook", "builder":
		dir = project.ScriptDir()
	case "lint", "test", "build":
		dir = project.TargetDir()
	}
	if len(dir) < 1 {
		return nil
	}
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		err = fmt.Errorf("failed to create dir:%w", err)
	}
	return err
}

var cleanAction boot.Action = func(project *boot.Project, _ string) error {
	keys := project.AllKeys()
	args := lo.FilterMap(keys, func(k string, _ int) (string, bool) {
		if strings.HasPrefix(k, "clean.") && project.GetBool(k) {
			return strings.Split(k, ".")[1], true
		} else {
			return "", false
		}
	})
	log.Printf("Cleaning project with flags: %s\n", strings.Join(args, ","))
	args = append([]string{"clean"}, args...)
	output, err := exec.Command("go", args...).CombinedOutput()
	msg := string(output)
	if err != nil {
		msg = color.RedString(string(output))
		log.Println(msg)
		return err //nolint:wrapcheck
	}
	log.Printf("Clean directory %s \n", project.TargetDir())
	err = filepath.WalkDir(project.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			err = os.Remove(path)
			if err != nil {
				log.Println(color.YellowString("failed to delete %s:%s", path, err.Error()))
			}
		} else if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err //nolint:wrapcheck
	})
	if err != nil {
		return fmt.Errorf("failed to delete %s :%w", project.TargetDir(), err)
	}
	return err //nolint:wrapcheck
}

var commitMsgAction boot.Action = func(project *boot.Project, cmd string) error {
	return validateCommitMsg(string(project.MsgPattern)) //nolint:wrapcheck
}

var lintAction boot.Action = func(project *boot.Project, cmd string) error {
	return newLinter().scan(project, project.IsCommitHook()) //nolint:wrapcheck
}

var testAction boot.Action = func(builder *boot.Project, cmd string) error {
	err := os.Chdir(builder.RootDir())
	if err != nil {
		return fmt.Errorf("failed to change directory:%w", err)
	}
	params := []string{"test", "-v", "-coverprofile", filepath.Join(builder.TargetDir(), testCoverOut), "./..."}
	// @todo for test parameter
	// if len(args) > 0 {
	//	params = append(params, args...)
	//}

	testCmd := exec.Command("go", params...)
	stdout, err := testCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	err = testCmd.Start()
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}
	scanner := bufio.NewScanner(stdout)
	// ok  	github.com/kcmvp/gob/builder	0.155s	coverage: 16.9% of statements
	pkr := regexp.MustCompile(`\sok\s+\S+\s+\S+s\s+coverage:\s+\S+% of statements`)
	// ?   	github.com/kcmvp/gob/infra	[no test files]
	ntr := regexp.MustCompile(`\s+\S+\s+\[no test files\]`)
	pkgCoverage := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "[build failed]"):
			line = color.RedString(line)
			err = errors.New("build failed") //nolint
		case strings.Contains(line, "--- FAIL:"), strings.Contains(line, "FAIL"):
			line = color.RedString(line)
			err = errors.New("test failure") //nolint
		case pkr.MatchString(line):
			line = color.GreenString(line)
			find := strings.Fields(pkr.FindString(line))
			pkgCoverage[find[1]] = find[4]
		case ntr.MatchString(line):
			line = color.YellowString(line)
			find := strings.Fields(ntr.FindString(line))
			pkgCoverage[find[0]] = "-"
		}
		log.Println(line)
	}
	if err != nil {
		return fmt.Errorf("test failed:%w", err)
	}

	data, err := json.MarshalIndent(&pkgCoverage, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal package coverage report:%w", err)
	}
	err = os.WriteFile(filepath.Join(builder.TargetDir(), testPackageCover), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save package coverage report:%w", err)
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(builder.TargetDir(), testCoverReport)
	params = []string{"tool", "cover", "-html", filepath.Join(builder.TargetDir(), testCoverOut), "-o", fileCover}
	out, err := exec.Command("go", params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%w", string(out), err)
	}
	log.Printf("coverage report is generated at %s \n", fileCover)
	return err //nolint:wrapcheck
}

var buildAction boot.Action = func(builder *boot.Project, cmd string) error {
	var targetFiles []string
	if len(targetFiles) == 0 {
		targetFiles = append(targetFiles, "main.go")
	}
	log.Println("build project ......")
	err := filepath.WalkDir(builder.RootDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		for _, t := range targetFiles {
			if strings.EqualFold(d.Name(), t) {
				if output, err := exec.Command("go", "build", "-o", builder.TargetDir(), path).CombinedOutput(); err != nil { //nolint
					log.Println(string(output))
					return err //nolint
				}
			}
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to build proejct:%w", err)
	}
	return err
}
