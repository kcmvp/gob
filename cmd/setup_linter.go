/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>

*/
package cmd

import (
	_ "embed"

	"github.com/kcmvp/gbt/builder/linter"
	"github.com/spf13/cobra"
)

//go:embed template/.golangci.yml
var golangCiTmp string

var version string

// linterCmd represents the linter command.
var linterCmd = &cobra.Command{
	Use:   "linter",
	Short: "setup linter for the project",
	Run: func(cmd *cobra.Command, args []string) {
		if ver, err := linter.ConfiguredVer(); err != nil {
			if v, err := linter.Install(version); err == nil {
				generateFile(golangCiTmp, linter.Cfg, v, false)
			}
		} else {
			linter.Install(ver)
		}
	},
}

func init() {
	setupCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", "latest", "golangci-lint version")
}
