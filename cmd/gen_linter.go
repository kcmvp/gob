/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"

	"github.com/kcmvp/gob/infra"
	"github.com/spf13/cobra"
)

var version string

// linterCmd represents the linter command.
var linterCmd = &cobra.Command{
	Use:   "linter",
	Short: "setup linter for the project",
	Run: func(cmd *cobra.Command, args []string) {
		if ver, err := infra.ConfiguredLinterVer(); err != nil {
			if v, err := infra.Install(version); err == nil {
				infra.GenerateLintCfg(v, false)
			}
		} else {
			infra.Install(ver)
		}
	},
}

func init() {
	setupCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", infra.LatestVer, "golangci-lint version")
}
