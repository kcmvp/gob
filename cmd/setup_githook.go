/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// githookCmd represents the githook command.
var githookCmd = &cobra.Command{
	Use:   boot.SetupHook.Name(),
	Short: "Setup git hook for project",
	Long:  `Setup git hooks for project, which include: commit_message, pre_push`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		return session.Run(builder.NewBuilder(), boot.SetupHook) //nolint
	},
}

func init() {
	setupCmd.AddCommand(githookCmd)
}
