/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// genCmd represents the setup command.
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate builder, git hook and other framework scaffold",
}

func init() {
	rootCmd.AddCommand(genCmd)
}
