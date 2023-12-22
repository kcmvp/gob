/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"github.com/fatih/color"
	"github.com/kcmvp/gb/cmd/root"
	"github.com/kcmvp/gb/cmd/shared"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"strings"
)

// onCommitMsg invoked by git commit-msg hook, an error returns when it fails to validate
// the commit message
var onCommitMsg shared.Execution = func(cmd *cobra.Command, args ...string) error {
	println("validate git commit message")
	return nil
}

var onCommit shared.Execution = func(cmd *cobra.Command, args ...string) error {
	root.BuildActions()
	return nil
}

var onPush shared.Execution = func(cmd *cobra.Command, args ...string) error {
	root.BuildActions()
	return nil
}

var execActions = []shared.CmdAction{
	{"commitMsg", onCommitMsg},
	{"onCommit", onCommit},
	{"onPush", onPush},
}

var actions = func() []string {
	return lo.Map(execActions, func(item shared.CmdAction, _ int) string {
		return item.A
	})
}()

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:       "exec",
	Short:     "Execute any tool that have been setup",
	Long:      `Execute any tool that have been setup`,
	ValidArgs: actions,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return errors.New(color.RedString(err.Error()))
		}
		if !lo.Contains(actions, args[0]) {
			return errors.New(color.RedString("valid args are: %s", strings.Join(actions, ",")))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := lo.Find(execActions, func(item shared.CmdAction) bool {
			return args[0] == item.A
		})
		return action.B(cmd, args...)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
