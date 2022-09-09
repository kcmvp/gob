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
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		boot.BindFlag(boot.SetupLinter, "version", cmd.Flags().Lookup("version"))
		return builder.NewBuilder(root).Run(boot.SetupLinter)
	},
}

func init() {
	setupCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
}
