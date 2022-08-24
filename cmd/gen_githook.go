/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gos/builder"
	"github.com/kcmvp/gos/infra"
	"github.com/spf13/cobra"
	"os"
)

// githookCmd represents the githook command.
var githookCmd = &cobra.Command{
	Use:   "githook",
	Short: "Generate git hook for project",
	Long:  `Generate git hooks for project, which include: commit_message, pre_push`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stat(git.GitDirName)
		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("project is not versioned in git: %w", err)
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		builder, _ := cmd.Context().Value(_ctxBuilder).(*builder.Builder)
		return infra.SetupHook(builder.ScriptDir(), builder.RootDir(), true)
	},
}

func init() {
	setupCmd.AddCommand(githookCmd)
}
