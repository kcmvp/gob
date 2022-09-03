/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// githookCmd represents the githook command.
var githookCmd = &cobra.Command{
	Use:   "githook",
	Short: "Generate git hook for project",
	Long:  `Generate git hooks for project, which include: commit_message, pre_push`,
	// PreRunE: func(cmd *cobra.Command, args []string) error {
	//	_, err := os.Stat(git.GitDirName)
	//	if errors.Is(err, os.ErrNotExist) {
	//		err = fmt.Errorf("project is not versioned in git: %w", err)
	//	}
	//	return err
	// },
	Run: func(cmd *cobra.Command, args []string) {
		// ctx := context.WithValue(cmd.Context(), builder.GenHook, true) //nolint
		builder.RunCtx(cmd.Context(), "gitHook")
	},
}

func init() {
	genCmd.AddCommand(githookCmd)
}
