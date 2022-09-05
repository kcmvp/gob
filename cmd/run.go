/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var (
	cleanCache bool
	testcache  bool
	modcache   bool
	fuzzcache  bool
)

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
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := os.Getwd()
		ctx := cmd.Context()
		builder := builder.NewBuilder(root)
		builder.BindPFlag("clean.-cache", cmd.Flags().Lookup("cache"))
		builder.BindPFlag("clean.-testcache", cmd.Flags().Lookup("testcache"))
		builder.BindPFlag("clean.-modcache", cmd.Flags().Lookup("modcache"))
		builder.BindPFlag("clean.-fuzzcache", cmd.Flags().Lookup("fuzzcache"))
		builder.RunCtx(ctx, args...)
	},
}

func init() {
	runCmd.Flags().BoolVarP(&cleanCache, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&testcache, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&modcache, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&fuzzcache, "fuzzcache", "f", false, "remove the entire module download cache")

	rootCmd.AddCommand(runCmd)
}
