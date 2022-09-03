/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/infra"
	"github.com/spf13/cobra"
)

var version string

// linterCmd represents the linter command.
var linterCmd = &cobra.Command{
	Use:   "linter",
	Short: "setup linter for the project",
	Run: func(cmd *cobra.Command, args []string) {
		linter := infra.NewLinter()
		if ver, err := linter.Configured(); err != nil {
			if v, err := linter.Install(version); err == nil {
				infra.GenLinterCfg(v, false)
			}
		} else {
			linter.Install(ver)
		}
	},
}

func init() {
	genCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
}
