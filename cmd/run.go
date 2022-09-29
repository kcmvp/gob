/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/kcmvp/gob/scaffolds"

	"github.com/kcmvp/gob/boot"
	"github.com/spf13/cobra"
)

var (
	cleanCache     bool
	cleanTestCache bool
	cleanModCache  bool
	cleanFuzzCache bool
	cleanDeleteAll bool
	lintFullScan   bool
	listAllRun     bool
)

const runCommand = "run"

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       runCommand,
	Short:     "Run build, test, lint and etc against current project, run `gob run -l` get more information",
	ValidArgs: scaffolds.ValidStack(runCommand),
	Args: func(cmd *cobra.Command, args []string) error {
		if listAllRun {
			return nil
		}
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		if err != nil {
			err = fmt.Errorf("run with %s against current project:%w", runCommand, err)
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if listAllRun {
			scaffolds.ListStack(runCommand)
			return nil
		}
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		session.BindFlag(boot.Clean, "-cache", cleanCache)
		session.BindFlag(boot.Clean, "-testcache", cleanTestCache)
		session.BindFlag(boot.Clean, "-modcache", cleanModCache)
		session.BindFlag(boot.Clean, "-fuzzcache", cleanFuzzCache)
		session.BindFlag(boot.Clean, "delete", cleanDeleteAll)
		session.BindFlag(boot.Lint, "all", lintFullScan)
		return session.Run(scaffolds.NewProject(), boot.ToCommands(args...)...) //nolint
	},
}

func init() {
	runCmd.Flags().BoolVarP(&listAllRun, "list", "l", false, "list all arguments for run command")
	runCmd.Flags().BoolVarP(&cleanCache, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&cleanTestCache, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&cleanModCache, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&cleanFuzzCache, "fuzzcache", "f", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&cleanDeleteAll, "delete", "d", false, "delete all the files in the target folder")
	runCmd.Flags().BoolVarP(&lintFullScan, "fullScan", "a", false, "Default only scan changed files, use -a to scan all files")

	rootCmd.AddCommand(runCmd)
}
