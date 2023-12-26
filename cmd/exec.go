/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/action"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

// validateCommitMsg invoked by git commit-msg hook, an error returns when it fails to validate
// the commit message
var validateCommitMsg action.Execution = func(cmd *cobra.Command, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf("please input commit message")
	}
	input, _ := os.ReadFile(args[1])
	regex := regexp.MustCompile(`\r?\n`)
	commitMsg := regex.ReplaceAllString(string(input), "")
	pattern, _ := lo.Last(args)
	regex = regexp.MustCompile(pattern)
	if !regex.MatchString(commitMsg) {
		return fmt.Errorf("error: commit message must follow %s", pattern)
	}
	return nil
}

var execValidArgs = func() []string {
	return lo.Map(internal.CurProject().Executions(), func(exec internal.Execution, _ int) string {
		return exec.CmdKey
	})
}()

func exec(execution internal.Execution, cmd *cobra.Command, args ...string) error {
	if execution.CmdKey == internal.CommitMsgCmd {
		args = append(args, execution.Actions...)
		return validateCommitMsg(cmd, args...)
	} else {
		for _, arg := range execution.Actions {
			fmt.Printf("start %s \n", arg)
			if err := action.Execute(cmd, arg); err != nil {
				return errors.New(color.RedString("failed to %s the project \n", arg))
			}
			color.Green("execute %s successfully", arg)
		}
		return nil
	}
}

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute any tool that have been setup",
	Long:  `Execute any tool that have been setup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
			return errors.New(color.RedString(err.Error()))
		}
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return errors.New(color.RedString(err.Error()))
		}
		if !lo.Contains(execValidArgs, args[0]) {
			return errors.New(color.RedString("invalid arg %s", args[0]))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		execution, _ := lo.Find(internal.CurProject().Executions(), func(exec internal.Execution) bool {
			return exec.CmdKey == args[0]
		})
		return exec(execution, cmd, args...)
	},
}

func init() {
	builderCmd.AddCommand(execCmd)
}
