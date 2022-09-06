/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// githookCmd represents the githook command.
var githookCmd = &cobra.Command{
	Use:   "githook",
	Short: "Setup git hook for project",
	Long:  `Setup git hooks for project, which include: commit_message, pre_push`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		return builder.NewBuilder(root).Run(cmd.Name())
	},
}

func init() {
	setupCmd.AddCommand(githookCmd)
}
