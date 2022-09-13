/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup builder, git hook and other framework scaffold",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
