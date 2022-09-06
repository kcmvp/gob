/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		ctx := cmd.Context()
		builder := builder.NewBuilder(root)
		builder.BindPFlag("clean.-cache", cmd.Flags().Lookup("cache"))
		builder.BindPFlag("clean.-testcache", cmd.Flags().Lookup("testcache"))
		builder.BindPFlag("clean.-modcache", cmd.Flags().Lookup("modcache"))
		builder.BindPFlag("clean.-fuzzcache", cmd.Flags().Lookup("fuzzcache"))
		// builder.BindPFlag("lint.new-from-rev", cmd.Flags().Lookup("new-from-rev"))
		// builder.BindPFlag("lint.fix", cmd.Flags().Lookup("fix"))
		return builder.RunCtx(ctx, args...)
	},
}

func init() {
	var boolValue bool
	// var stringValue string
	runCmd.Flags().BoolVarP(&boolValue, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&boolValue, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&boolValue, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&boolValue, "fuzzcache", "f", false, "remove the entire module download cache")

	// runCmd.Flags().StringVarP(&stringValue, "new-from-head", "n", true, "only scan issues from new code(changed code)")
	// runCmd.Flags().StringVar(&fuzzcache, "fix", "x", true, "only scan issues from new code(changed code)")

	rootCmd.AddCommand(runCmd)
}
