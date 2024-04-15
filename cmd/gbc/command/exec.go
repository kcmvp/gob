/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package command

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const pushDeleteHash = "0000000000000000000000000000000000000000"

// validateCommitMsg invoked by git commit-msg hook, an error returns when it fails to validate
// the commit message
var validateCommitMsg Execution = func(_ *cobra.Command, args ...string) error {
	if len(args) < 2 {
		return fmt.Errorf(color.RedString("please input commit message"))
	}
	input, _ := os.ReadFile(args[1])
	regex := regexp.MustCompile(`\r?\n`)
	commitMsg := regex.ReplaceAllString(string(input), "")
	pattern, _ := lo.Last(args)
	regex = regexp.MustCompile(pattern)
	if !regex.MatchString(commitMsg) {
		return fmt.Errorf(color.RedString("commit message must follow %s", pattern))
	}
	return nil
}

func execValidArgs() []string {
	return lo.Map(artifact.CurProject().Executions(), func(exec artifact.Execution, _ int) string {
		return exec.CmdKey
	})
}

func pushDelete(cmd string) bool {
	if cmd == artifact.PrePush {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, pushDeleteHash) && strings.Contains(line, "delete") {
				return true
			}
		}
	}
	return false
}

func do(execution artifact.Execution, cmd *cobra.Command, args ...string) error {
	if execution.CmdKey == artifact.CommitMsg {
		args = append(args, execution.Actions...)
		return validateCommitMsg(nil, args...)
	}
	if pushDelete(execution.CmdKey) {
		return nil
	}
	// process hook
	for _, arg := range execution.Actions {
		if err := execute(cmd, arg); err != nil {
			return errors.New(color.RedString("failed to %s the project \n", arg))
		}
	}
	return nil
}

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute any tools that have been setup",
	Long:  `Execute any tools that have been setup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(3)(cmd, args); err != nil {
			return errors.New(color.RedString(err.Error()))
		}
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return errors.New(color.RedString(err.Error()))
		}
		if !lo.Contains(execValidArgs(), args[0]) {
			return errors.New(color.RedString("invalid arg %s", args[0]))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		execution, _ := lo.Find(artifact.CurProject().Executions(), func(exec artifact.Execution) bool {
			return exec.CmdKey == args[0]
		})
		return do(execution, cmd, args...)
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
