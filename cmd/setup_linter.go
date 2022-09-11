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
	Use:   boot.SetupLinter.Name(),
	Short: "setup linter for the project",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		boot.BindFlag(boot.SetupLinter, "version", version)
		return boot.Run(builder.NewBuilder(root), boot.SetupLinter)
	},
}

func init() {
	linterCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
	setupCmd.AddCommand(linterCmd)
}
