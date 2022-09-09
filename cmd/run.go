/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/kcmvp/gob/boot"
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
		builder := builder.NewBuilder(root)
		boot.BindFlag(boot.Clean, "-cache", cmd.Flags().Lookup("cache"))
		boot.BindFlag(boot.Clean, "-testcache", cmd.Flags().Lookup("testcache"))
		boot.BindFlag(boot.Clean, "-modcache", cmd.Flags().Lookup("modcache"))
		boot.BindFlag(boot.Clean, "-fuzzcache", cmd.Flags().Lookup("fuzzcache"))
		boot.BindFlag(boot.Lint, "all", cmd.Flags().Lookup("scanAll"))
		return builder.Run(boot.ToCommand(args...)...)
	},
}

func init() {
	var boolValue bool

	runCmd.Flags().BoolVarP(&boolValue, "cache", "c", false, "remove the entire go build cache")
	runCmd.Flags().BoolVarP(&boolValue, "testcache", "t", false, "expire all test results")
	runCmd.Flags().BoolVarP(&boolValue, "modcache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&boolValue, "fuzzcache", "f", false, "remove the entire module download cache")

	runCmd.Flags().BoolVarP(&boolValue, "scanAll", "a", false, "Default only scan changed files, use -a to scan all files")

	rootCmd.AddCommand(runCmd)
}
