/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var cleanCache bool
var cleanTestCache bool
var cleanModCache bool
var clanFuzzCache bool
var lintScanAll bool

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       "run",
	Short:     "Run 'clean', 'test', 'lint', 'build' commands against current project",
	ValidArgs: []string{"clean", "test", "lint", "build"},
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		builder := builder.NewBuilder()
		boot.BindFlag(boot.Clean, "-cache", cleanCache)
		boot.BindFlag(boot.Clean, "-testcache", cleanTestCache)
		boot.BindFlag(boot.Clean, "-modcache", cleanModCache)
		boot.BindFlag(boot.Clean, "-fuzzcache", clanFuzzCache)
		boot.BindFlag(boot.Lint, "all", lintScanAll)
		return boot.Run(builder, boot.ToCommands(args...)...)
	},
}

func init() {

	runCmd.Flags().BoolVarP(&cleanCache, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&cleanTestCache, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&cleanModCache, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&clanFuzzCache, "fuzzcache", "f", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&lintScanAll, "scanAll", "a", false, "Default only scan changed files, use -a to scan all files")

	rootCmd.AddCommand(runCmd)
}
