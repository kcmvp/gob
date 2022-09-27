/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print out gob version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("version: v1.0.2")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
