package shared

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Execution func(cmd *cobra.Command, args ...string) error

type CmdAction lo.Tuple2[string, Execution]

func PrintCmd(cmd *cobra.Command, msg string) error {
	if ok, file := internal.TestEnv(); ok {
		// Get the call stack
		outputFile, err := os.Create(filepath.Join(internal.CurProject().Target(), file))
		if err != nil {
			return err
		}
		defer outputFile.Close()
		writer := io.MultiWriter(os.Stdout, outputFile)
		fmt.Fprintln(writer, msg)
	} else {
		cmd.Println(msg)
	}
	return nil
}

func StreamExtCmdOutput(cmd *exec.Cmd, file string, errWords ...string) error {
	// Create a pipe to capture the command's combined output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	outputFile, err := os.Create(file)
	if err != nil {
		return err
	}
	defer outputFile.Close()
	writer := io.MultiWriter(os.Stdout, outputFile)
	err = cmd.Start()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdout)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if lo.SomeBy(errWords, func(item string) bool {
				return strings.Contains(line, item)
			}) {
				color.Red(line)
				fmt.Fprintln(outputFile, line)
			} else {
				fmt.Fprintln(writer, line)
			}
		}
	}()
	// Wait for the command to finish
	return cmd.Wait()
}

// embed commit_msg_hook.tmpl
var hook []byte

var InitGitHook = func() error {
	fmt.Printf("Initialize git hook ......\n")
	if _, err := git.PlainOpen(internal.CurProject().Root()); err != nil {
		color.Yellow("Project is not in the source control, please add it to source repository")
		return err
	}
	script := filepath.Join(internal.CurProject().Root(), "commit_msg_hook.go")
	err := os.WriteFile(script, hook, os.ModePerm)
	if err != nil {
		return err
	}
	hookMap := map[string]string{
		//"commit-msg": fmt.Sprintf("go run %s $1 $2", filepath.Join(internal.CurProject().Root(), "commit_msg_hook.go")),
		"commit-msg": "gb exec gmh",
		"pre-commit": "gb lint test",
		"pre-push":   "gb lint test",
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
