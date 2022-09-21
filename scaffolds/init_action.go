package scaffolds

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
	"github.com/thedevsaddam/gojsonq/v2"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/boot"
)

//go:embed template/stack.json
var stacks string

var initBuilder boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
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

var initHook boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
	log.Println("Setup git hooks")
	err := initGitHooks(project.GitHome(), project.ScriptDir())
	var gitErr *GitErr
	if errors.As(err, &gitErr) && command != boot.InitHook {
		log.Println(color.YellowString("Project is not in the git repository"))
	} else if err != nil {
		err = fmt.Errorf("failed to setup hook:%w", err)
	} else if command == boot.InitHook {
		log.Println("git hooks are setup successfully")
	}
	return err
}

var initLinter boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
	log.Println("Setup linters")
	linter := newLinter()
	version := session.GetFlagString(command, "version")
	cfgVersion := project.Config().GetString(linter.CfgVerKey())
	if cfgVersion != version {
		version = cfgVersion
	}
	// to get the real version
	version, err := linter.Install(version)
	if err != nil {
		return err //nolint
	}
	err = boot.GenerateFile(golangCiTmp, filepath.Join(project.RootDir(), lintCfg), nil, false)
	if err != nil {
		return fmt.Errorf("failed to generate lint config:%w", err)
	}
	project.SaveConfig(linter.CfgVerKey(), version)
	return nil
}

var listStacks boot.Action = func(session *boot.Session, project boot.Project, command boot.Command) error {
	jq := gojsonq.New().FromString(stacks).From("stacks")
	ct := table.Table{}
	ct.SetTitle("Supported scaffolds")
	ct.AppendHeader(table.Row{"#", "Name", "Category", "Module", "Description"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	ct.SetStyle(style)
	data := jq.Get()
	consoleRows := lo.Map(data.([]interface{}), func(t interface{}, i int) table.Row {
		tm := t.(map[string]interface{})
		return table.Row{i + 1, tm["name"], tm["category"], tm["module"], tm["Description"]}
	})
	ct.AppendRows(consoleRows)
	ct.SortBy([]table.SortBy{{Name: "Category", Mode: table.Asc}})
	fmt.Println(ct.Render())
	return nil
}
