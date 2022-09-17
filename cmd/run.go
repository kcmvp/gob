/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var (
	cleanCache     bool
	cleanTestCache bool
	cleanModCache  bool
	cleanFuzzCache bool
	cleanDeleteAll bool
	lintFullScan   bool
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       "run",
	Short:     "Run 'clean', 'test', 'lint', 'build' and 'report' commands against current project",
	ValidArgs: []string{"clean", "test", "lint", "build", "report"},
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		session.BindFlag(boot.Clean, "-cache", cleanCache)
		session.BindFlag(boot.Clean, "-testcache", cleanTestCache)
		session.BindFlag(boot.Clean, "-modcache", cleanModCache)
		session.BindFlag(boot.Clean, "-fuzzcache", cleanFuzzCache)
		session.BindFlag(boot.Clean, "delete", cleanDeleteAll)
		session.BindFlag(boot.Lint, "all", lintFullScan)
		return session.Run(builder.NewBuilder(), boot.ToCommands(args...)...) //nolint
	},
}

func init() {
	runCmd.Flags().BoolVarP(&cleanCache, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&cleanTestCache, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&cleanModCache, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&cleanFuzzCache, "fuzzcache", "f", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&cleanDeleteAll, "delete", "d", false, "delete all the files in the target folder")
	runCmd.Flags().BoolVarP(&lintFullScan, "fullScan", "a", false, "Default only scan changed files, use -a to scan all files")

	rootCmd.AddCommand(runCmd)
}
