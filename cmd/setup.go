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
	Short: "setup 'builder', 'githook', 'linter' and other framework scaffold",
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
