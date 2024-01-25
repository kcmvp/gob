/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/samber/lo"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup scaffold of Go project",
	Long:  `Setup scaffold for most common used frameworks`,
	ValidArgs: func() []string {
		return lo.Map(setupActions(), func(action Action, _ int) string {
			return action.A
		})
	}(),
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		action, _ := lo.Find(setupActions(), func(action Action) bool {
			return action.A == args[0]
		})
		_ = action.B(cmd, args...)
	},
}

func init() {
	builderCmd.AddCommand(setupCmd)
}
