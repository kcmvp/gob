// Package cmd /*
package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

//go:embed resources/*
var resources embed.FS

const resourceDir = "resources"

// builderCmd represents the base command when called without any subcommands
var builderCmd = &cobra.Command{
	Use:   "gob",
	Short: "Go project boot",
	Long:  `Supply most frequently used tool and best practices for go project development`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return validBuilderArgs(), cobra.ShellCompDirectiveError
	},
	Args: func(cmd *cobra.Command, args []string) error {
		if !lo.Every(validBuilderArgs(), args) {
			return fmt.Errorf("valid args are : %s", validBuilderArgs())
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return internal.CurProject().Validate()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range lo.Uniq(args) {
			if err := execute(cmd, arg); err != nil {
				return errors.New(color.RedString("%s \n", err.Error()))
			}
		}
		return nil
	},
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
	builderCmd.Flags().BoolVar(&CleanCache, cleanCacheFlag, false, "to remove the entire go build cache")
	builderCmd.Flags().BoolVar(&CleanTestCache, cleanTestCacheFlag, false, "to expire all test results in the go build cache")
	builderCmd.Flags().BoolVar(&CleanModCache, cleanModCacheFlag, false, "to remove the entire module download cache")
}
