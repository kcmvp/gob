/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"

	"github.com/kcmvp/gob/builder"

	"github.com/kcmvp/gob/infra"
	"github.com/spf13/cobra"
)

var version string

// linterCmd represents the linter command.
var linterCmd = &cobra.Command{
	Use:   "linter",
	Short: "setup linter for the project",
	Run: func(cmd *cobra.Command, args []string) {
		if ver, err := infra.GetLinterVer(builder.GetBuilder(cmd.Context()).RootDir()); err != nil {
			if v, err := infra.InstallLinter(version); err == nil {
				infra.GenLinterCfg(v, false)
			}
		} else {
			infra.InstallLinter(ver)
		}
	},
}

func init() {
	genCmd.AddCommand(linterCmd)
	linterCmd.Flags().StringVarP(&version, "version", "v", infra.LatestVer, "golangci-lint version")
}
