// Package cmd /*
package cmd

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gb/cmd/action"
	"github.com/kcmvp/gb/cmd/root"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gb",
	Short: "Go project boot",
	Long:  `Supply most frequently used tool and best practices for go project development`,
	ValidArgs: lo.Map(root.BuildActions(), func(item action.CmdAction, _ int) string {
		return item.A
	}),
	Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MinimumNArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		return buildProject(cmd, args)
	},
}

func Execute() {
	currentDir, err := os.Getwd()
	if err != nil {
		color.Red("Failed to execute command: %v", err)
		return
	}
	if internal.CurProject().Root() != currentDir {
		color.Red("Please execute the command in the project root dir")
		return
	}
	ctx := context.Background()
	if err = rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func buildProject(cmd *cobra.Command, args []string) error {
	uArgs := lo.Uniq(args)
	expectedActions := lo.Filter(root.BuildActions(), func(item action.CmdAction, index int) bool {
		return lo.Contains(uArgs, item.A)
	})
	// Check if the folder exists
	os.Mkdir(internal.CurProject().Target(), os.ModePerm)
	for _, action := range expectedActions {
		msg := fmt.Sprintf("Start %s project", action.A)
		fmt.Printf("%-20s ...... \n", msg)
		if err := action.B(cmd); err != nil {
			color.Red("Failed to %s project %v \n", action.A, err.Error())
			return err
		}
	}
	return nil
}

func init() {
	rootCmd.SetErrPrefix(color.RedString("Error:"))
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVar(&root.CleanCache, root.CleanCacheFlag, false, "to remove the entire go build cache")
	rootCmd.Flags().BoolVar(&root.CleanTestCache, root.CleanTestCacheFlag, false, "to expire all test results in the go build cache")
	rootCmd.Flags().BoolVar(&root.CleanModCache, root.CleanModCacheFlag, false, "to remove the entire module download cache")
	rootCmd.Flags().BoolVar(&root.LintAll, root.LintAllFlag, false, "lint scan all source code, default only on changed source code")
}
