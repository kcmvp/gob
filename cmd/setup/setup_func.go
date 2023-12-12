package setup

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kcmvp/gob/cmd/action"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

//go:embed setup.json
var data []byte

//go:embed commit_msg.tmpl
var hook []byte

var setups []Setup

type Setup struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func init() {
	json.Unmarshal(data, &setups)
	// for each command there should be a action
	if missed, ok := lo.Find(setups, func(setup Setup) bool {
		return !lo.ContainsBy(Actions, func(action action.CmdAction) bool {
			return action.A == setup.Name
		})
	}); ok {
		log.Fatal(color.RedString("Fatal: missed action for %s\n", missed.Name))
	}
}

var Actions = []action.CmdAction{
	{"list", list},
	{"githook", gitHook},
	{"gitflow", gitflow},
	{"onion", onion},
}

var gitflow action.Execution = func(cmd *cobra.Command, args ...string) error {
	return nil
}

var onion action.Execution = func(cmd *cobra.Command, args ...string) error {
	return nil
}

var list action.Execution = func(cmd *cobra.Command, args ...string) error {
	ct := table.Table{}
	ct.SetTitle("Available Actions")
	ct.AppendRow(table.Row{"#", "Name", "Description"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.Title.Align = text.AlignCenter
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	consoleRows := lo.Map(setups, func(setup Setup, i int) table.Row {
		return table.Row{i + 1, setup.Name, setup.Desc}
	})
	ct.AppendRows(consoleRows)
	action.PrintCmd(cmd, ct.Render())
	return nil
}

var gitHook action.Execution = func(cmd *cobra.Command, args ...string) error {
	if _, err := git.PlainOpen(internal.CurProject().Root()); err != nil {
		color.Yellow("Project is not in the source control, please add it to source repository")
		return err
	}
	config := filepath.Join(internal.CurProject().Root(), "config")
	os.Mkdir(config, os.ModePerm)
	err := os.WriteFile(filepath.Join(config, "commit_msg.go"), hook, os.ModePerm)
	if err != nil {
		return err
	}
	hookMap := map[string]string{
		"commit-msg": fmt.Sprintf("go run %s/config/commit_msg.go $1 $2", internal.CurProject().Root()),
		"pre-commit": "gob lint test",
		"pre-push":   "gob lint test",
	}
	shell := lo.IfF(internal.Windows(), func() string {
		return "#!/usr/bin/env pwsh\n"
	}).Else("#!/bin/sh\n")
	for name, script := range hookMap {
		msgHook, _ := os.OpenFile(filepath.Join(internal.CurProject().Root(), ".git", "hooks", name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		writer := bufio.NewWriter(msgHook)
		writer.WriteString(shell)
		writer.WriteString("\n")
		writer.WriteString(script)
		writer.Flush()
		msgHook.Close()
	}
	return nil
}
