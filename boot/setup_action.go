package boot

import (
	_ "embed" //nolint
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/mod/modfile"

	"github.com/fatih/color"
)

var setupBuilder Action = func(session *Session, project *Project, command Command) error {
	log.Println("Creating project build file")
	var err error
	var tf []byte
	tf, err = templateDir.ReadFile(filepath.Join("template", "builder.tmpl"))
	if err == nil {
		err = GenerateFile(string(tf), filepath.Join(project.ScriptDir(), "builder.go"), nil, false)
	}
	if err != nil {
		err = fmt.Errorf("failed to generate builder script:%w", err)
	}
	return err
}

var setupHook Action = func(session *Session, project *Project, command Command) error {
	log.Println("Setup git hooks")
	err := setupGitHooks(project)
	var gitErr *GitErr
	if errors.As(err, &gitErr) && command != SetupHook {
		log.Println(color.YellowString("Project is not in the git repository"))
	} else if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if command == SetupHook {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var setupLinter Action = func(session *Session, project *Project, command Command) error {
	log.Println("Setup linters")
	linter := newLinter()
	version := session.GetFlagString(command, "version")
	cfgVersion := project.Config().GetString(linter.Cmd())
	if len(cfgVersion) > 0 && cfgVersion != version {
		version = cfgVersion
	}
	// to get the real version
	version, err := linter.Install(version)
	if err != nil {
		return err //nolint
	}
	err = GenerateFile(golangCiTmp, filepath.Join(project.RootDir(), lintCfg), nil, false)
	if err != nil {
		return fmt.Errorf("failed to generate lint config:%w", err)
	}
	project.SaveConfig(linter.Cmd(), version)
	return nil
}

var setupGitFlow Action = func(session *Session, project *Project, command Command) error {
	log.Println("Setup Github Flow")
	abs, _ := filepath.Abs(filepath.Join(project.RootDir(), ".github", "workflows", "build.yml"))
	if _, err := os.Stat(abs); errors.Is(err, os.ErrNotExist) {
		tf, err := templateDir.ReadFile(filepath.Join("template", "build.workflow.tmpl"))
		if err == nil {
			err = os.WriteFile(abs, tf, os.ModePerm)
		}
		return err // nolint
	}
	return nil
}

var getGob Action = func(session *Session, project *Project, command Command) error {
	gob := "github.com/kcmvp/gob"
	_, ok := lo.Find(project.mod.Require, func(r *modfile.Require) bool {
		return strings.HasPrefix(r.Mod.String(), gob)
	})
	if !ok {
		cmd := exec.Command("go", "get", fmt.Sprintf("%s@latest", gob)) //nolint
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
		}
	}
	return nil
}
