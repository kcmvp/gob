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
	Use:       "gob",
	Short:     "Go project boot",
	Long:      `Go pluggable toolchain and best practice`,
	ValidArgs: validBuilderArgs(),
	Args: func(cmd *cobra.Command, args []string) error {
		if !lo.Every(validBuilderArgs(), args) {
			return fmt.Errorf(color.RedString("valid args are : %s", validBuilderArgs()))
		}
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return fmt.Errorf(color.RedString(err.Error()))
		}
		return nil
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		internal.CurProject().Validate()
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

func Execute() error {
	currentDir, _ := os.Getwd()
	if internal.CurProject().Root() != currentDir {
		return fmt.Errorf(color.RedString("Please execute the command in the project root dir"))
	}
	ctx := context.Background()
	if err := builderCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf(color.RedString(err.Error()))
	}
	return nil
}

func init() {
	builderCmd.SetErrPrefix(color.RedString("Error:"))
	builderCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	builderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
