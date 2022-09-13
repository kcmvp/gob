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
	"time"

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

var genBuilder boot.Action = func(project boot.Project, command boot.Command) error {
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

var genHook boot.Action = func(project boot.Project, command boot.Command) error {
	log.Println("Setup git hooks")
	err := genGitHooks(project.GitHome(), project.ScriptDir())
	if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if command.Name() == "githook" {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var setupLinter boot.Action = func(project boot.Project, command boot.Command) error {
	log.Println("Setup linters")
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

var createDirAction boot.Action = func(project boot.Project, command boot.Command) error {
	log.Println("Creating project directories")
	var dir string
	switch command.Name() {
	case boot.SetupHook.Name(), boot.SetupBuilder.Name():
		dir = project.ScriptDir()
	case boot.Lint.Name(), boot.Test.Name(), boot.Build.Name():
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

var cleanAction boot.Action = func(project boot.Project, command boot.Command) error {
	log.Println("Cleaning project")
	flags := lo.FilterMap(command.ValidFlags(), func(flag string, i int) (string, bool) {
		return flag, boot.GetFlag[bool](command, flag)
	})
	if len(flags) > 0 {
		log.Printf("Flags: %s\n", strings.Join(flags, ","))
		args := append([]string{command.Name()}, flags...)
		output, err := exec.Command("go", args...).CombinedOutput()
		msg := string(output)
		if err != nil {
			msg = color.RedString(string(output))
			log.Println(msg)
			return err //nolint:wrapcheck
		}
		boot.SaveExecCtx(command, fmt.Sprintf("%s %s", "go", strings.Join(args, " ")))
	}
	err := filepath.WalkDir(project.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasSuffix(d.Name(), ".tmp") {
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

var commitMsgAction boot.Action = func(project boot.Project, command boot.Command) error {
	builder := project.(*Builder)
	log.Println("Validate commit message")
	return validateCommitMsg(string(builder.MsgPattern)) //nolint:wrapcheck
}

var lintAction boot.Action = func(project boot.Project, command boot.Command) error {
	builder := project.(*Builder)
	log.Println("Running linters against source code")
	return newLinter().scan(builder, command) //nolint:wrapcheck
}

var testAction boot.Action = func(builder boot.Project, command boot.Command) error {
	err := os.Chdir(builder.RootDir())
	log.Println("Running unit test")
	if err != nil {
		return fmt.Errorf("failed to change directory:%w", err)
	}
	suffix := fmt.Sprintf(".%d.tmp", time.Now().UnixMilli())
	params := []string{"test", "-v", "-coverprofile", filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", testCoverOut, suffix)), "./..."}

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
	err = os.WriteFile(filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", testPackageCover, suffix)), data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save package coverage report:%w", err)
	}
	//  go tool cover -func ./targetDir/coverage.data
	fileCover := filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", testCoverReport, suffix))
	params = []string{"tool", "cover", "-html", filepath.Join(builder.TargetDir(), fmt.Sprintf("%s%s", testCoverOut, suffix)), "-o", fileCover}
	out, err := exec.Command("go", params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s:%w", string(out), err)
	}
	log.Printf("coverage report is generated at %s \n", strings.TrimSuffix(fileCover, suffix))
	rename(builder.TargetDir(), suffix)
	return err //nolint:wrapcheck
}

func rename(dir, suffix string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), suffix) {
			np := strings.TrimSuffix(path, suffix)
			err = os.Rename(path, np)
			if err != nil {
				log.Println(color.YellowString("failed to rename file %s to %s: ", path, np, err.Error()))
			}
		} else if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err //nolint:wrapcheck
	})
}

var buildAction boot.Action = func(builder boot.Project, command boot.Command) error {
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
