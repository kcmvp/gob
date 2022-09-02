/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var scanNew = false
var cleanCache = false
var testCache = false
var modCache = false
var fuzzcache = false

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       "run",
	Short:     fmt.Sprintf("Valid run flags are: %s", strings.Join(builder.Children("run"), ",")),
	ValidArgs: builder.Children("run"),
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Executing commands :%s\n", strings.Join(args, ","))
		funFlags := map[string]bool{}
		funFlags["-n"] = scanNew
		funFlags["-cache"] = cleanCache
		funFlags["-testcache"] = testCache
		funFlags["-modcache"] = modCache
		funFlags["-fuzzcache"] = fuzzcache
		ctx := context.WithValue(cmd.Context(), builder.CtxKeyRunFlags, funFlags)
		builder.RunCtx(ctx, args...)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&scanNew, "new", "n", true, " Show only new lint issues (default)")
	runCmd.Flags().BoolVarP(&cleanCache, "clean-cache", "c", false, "remove the entire go build cache.")
	runCmd.Flags().BoolVarP(&testCache, "clean-testcache", "t", false, "expire all test results in the go build cache")
	runCmd.Flags().BoolVarP(&modCache, "clean-modecache", "m", false, "remove the entire module download cache")
	runCmd.Flags().BoolVarP(&fuzzcache, "clean-fuzzcache", "f", false, "remove files build cache for fuzz testing")
}
