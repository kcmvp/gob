/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"os"

	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var version string

// linterCmd represents the linter command.
var linterCmd = &cobra.Command{
	Use:   "linter",
	Short: "setup linter for the project",
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := os.Getwd()
		project := builder.NewBuilder(root)
		project.BindPFlag("version", cmd.Flags().Lookup("version"))
		project.Run(cmd.Name())
	},
}

func init() {
	genCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
}
