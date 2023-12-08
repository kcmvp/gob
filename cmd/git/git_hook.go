package git

import (
	"bufio"
	"embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gob/cmd/common"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

//go:embed *.tmpl
var gitHookTempDir embed.FS

var SetupGitHookFunc common.ArgFunc = func(cmd *cobra.Command) error {
	// check the project in git nor not
	if _, err := git.PlainOpen(internal.CurProject().Root()); err != nil {
		color.Yellow("Project is not in the source control, please add it to source repository")
		return err
	}
	// generate commit_msg hook
	os.Mkdir(filepath.Join(internal.CurProject().Root(), "config"), 0755)
	src, _ := gitHookTempDir.Open("commit_msg.tmpl")
	defer src.Close()
	dest, _ := os.Create(filepath.Join(internal.CurProject().Root(), "config", "commit_msg.go"))
	defer dest.Close()
	io.Copy(dest, src)

	hookMap := map[string]string{
		"commit-msg": fmt.Sprintf("go run %s/config/commit_msg.go $1 $2", internal.CurProject().Root()),
		"pre-commit": "gob lint test",
		"pre-push":   "gob lint test",
	}
	shell := lo.IfF(runtime.GOOS == "windows", func() string {
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

var Validate = func(cmd *cobra.Command, args []string) error {
	return nil
}
