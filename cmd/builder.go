// Package cmd /*
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gb/cmd/builder"
	"github.com/kcmvp/gb/cmd/shared"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

// builderCmd represents the base command when called without any subcommands
var builderCmd = &cobra.Command{
	Use:   "gb",
	Short: "Go project boot",
	Long:  `Supply most frequently used tool and best practices for go project development`,
	ValidArgs: lo.Map(builder.Actions(), func(item shared.CmdAction, _ int) string {
		return item.A
	}),
	Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MinimumNArgs(1)),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		internal.CurProject().Setup(false)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := build(cmd, args); err != nil {
			return errors.New(color.RedString("%s \n", err.Error()))
		}
		return nil
	},
}

func build(cmd *cobra.Command, args []string) error {
	actions := lo.Filter(builder.Actions(), func(action shared.CmdAction, _ int) bool {
		return lo.Contains(args, action.A)
	})
	for _, action := range actions {
		msg := fmt.Sprintf("Start %s project", action.A)
		fmt.Printf("%-20s ...... \n", msg)
		if err := action.B(cmd, action.A); err != nil {
			return err
		}
	}
	return nil
}

func Execute() {
	currentDir, err := os.Getwd()
	if err != nil {
		color.Red("Failed to execute command: %v", err)
		os.Exit(1)
	}
	if internal.CurProject().Root() != currentDir {
		color.Red("Please execute the command in the project root dir")
		os.Exit(1)
	}
	ctx := context.Background()
	if err = builderCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	builderCmd.SetErrPrefix(color.RedString("Error:"))
	builderCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	builderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	builderCmd.Flags().BoolVar(&builder.CleanCache, builder.CleanCacheFlag, false, "to remove the entire go build cache")
	builderCmd.Flags().BoolVar(&builder.CleanTestCache, builder.CleanTestCacheFlag, false, "to expire all test results in the go build cache")
	builderCmd.Flags().BoolVar(&builder.CleanModCache, builder.CleanModCacheFlag, false, "to remove the entire module download cache")
	builderCmd.Flags().BoolVar(&builder.LintAll, builder.LintAllFlag, false, "lint scan all source code, default only on changed source code")
}