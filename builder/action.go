package builder

import (
	"bufio"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samber/lo"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/boot"
)

const (
	testCoverOut     = "cover.out"
	testCoverReport  = "cover.html"
	testPackageCover = "cover.json"
)

//go:embed template/*.tmpl
var templateDir embed.FS

var GenBuilder boot.Action = func(project boot.Project, command boot.Command) error {
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

var GenHook boot.Action = func(project boot.Project, command boot.Command) error {
	err := genGitHooks(project.GitHome(), project.ScriptDir())
	if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if command.Name() == "githook" {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var SetupLinter boot.Action = func(project boot.Project, command boot.Command) error {
	linter := newLinter()

	version := boot.GetFlag[string](command, "version")
	cfgVersion := project.Config().GetString(linter.CfgVerKey())
	if cfgVersion != version {
		version = cfgVersion
	}
	// to get the real version
	version, err := linter.Install(version)
	if err != nil {
		return err
	}
	err = boot.GenerateFile(golangCiTmp, lintCfg, nil, false)
	if err != nil {
		return fmt.Errorf("failed to generate lint config:%w", err)
	}
	project.SaveConfig(linter.CfgVerKey(), version)
	return nil
}

var CreateDirAction boot.Action = func(project boot.Project, command boot.Command) error {
	var dir string
	// todo fix the bug
	switch command.Name() {
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

var CleanAction boot.Action = func(project boot.Project, command boot.Command) error {
	flags := lo.FilterMap(command.ValidFlags(), func(flag string, i int) (string, bool) {
		return flag, boot.GetFlag[bool](command, flag)
	})
	log.Printf("Cleaning project with flags: %s\n", strings.Join(flags, ","))
	args := append([]string{command.Name()}, flags...)
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

var CommitMsgAction boot.Action = func(project boot.Project, command boot.Command) error {
	builder := project.(*Builder)
	return validateCommitMsg(string(builder.MsgPattern)) //nolint:wrapcheck
}

var LintAction boot.Action = func(project boot.Project, command boot.Command) error {
	builder := project.(*Builder)
	return newLinter().scan(builder, command) //nolint:wrapcheck
}

var TestAction boot.Action = func(builder boot.Project, command boot.Command) error {
	err := os.Chdir(builder.RootDir())
	if err != nil {
		return fmt.Errorf("failed to change directory:%w", err)
	}
	params := []string{"test", "-v", "-coverprofile", filepath.Join(builder.TargetDir(), testCoverOut), "./..."}

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

var BuildAction boot.Action = func(builder boot.Project, command boot.Command) error {
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
