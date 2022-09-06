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
	Short: "Generate git hook for project",
	Long:  `Generate git hooks for project, which include: commit_message, pre_push`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		return builder.NewBuilder(root).Run(cmd.Name())
	},
}

func init() {
	genCmd.AddCommand(githookCmd)
}
