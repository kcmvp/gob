// Package cmd /*
package cmd

import (
	"context"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

var validArgsFun = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var validArgs []string
	if cmd.Name() == "gob" {
		validArgs = append(validArgs, []string{"action", "clean", "test", "package"}...)
	}
	return validArgs, cobra.ShellCompDirectiveNoSpace
}

const (
	CleanCacheFlag     = "cache"
	CleanTestCacheFlag = "testcache"
	CleanModCacheFlag  = "modcache"
)

// cache the same as 'go clean -cache'
var cache bool

// testCache the same as `go clean -testcache'
var testCache bool

// modCache the same as 'go clean -modcache'
var modCache bool

// report generate test or lint report
var report bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: "Go project boot",
	Long:  `Supply most frequently used tool and best practices for go project development`,
	ValidArgs: lo.Map(buildActions, func(item lo.Tuple2[string, buildCmdFunc], _ int) string {
		return item.A
	}),
	Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		buildProject(cmd, args)
	},
}

func Execute() {
	currentDir, err := os.Getwd()
	if err != nil {
		internal.Red.Printf("Failed to execute command: %v", err)
		return
	}
	if internal.CurProject().Root() != currentDir {
		internal.Yellow.Println("Please execute the command in the project root dir")
		return
	}
	ctx := context.Background()
	if err = rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetErrPrefix(internal.Red.Sprintf("Error:"))
	rootCmd.AddCommand(setupCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolVar(&cache, CleanCacheFlag, false, "to remove the entire go build cache")
	rootCmd.Flags().BoolVar(&testCache, CleanTestCacheFlag, false, "to expire all test results in the go build cache")
	rootCmd.Flags().BoolVar(&modCache, CleanModCacheFlag, false, "to remove the entire module download cache")
}
