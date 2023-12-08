// Package cmd /*
package cmd

import (
	"context"
	"github.com/kcmvp/gob/cmd/common"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

const (
	CleanCacheFlag     = "cache"
	CleanTestCacheFlag = "testcache"
	CleanModCacheFlag  = "modcache"
	LintAllFlag        = "all"
)

// cleanCache the same as 'go clean -cache'
var cleanCache bool

// cleanTestCache the same as `go clean -testcache'
var cleanTestCache bool

// cleanModCache the same as 'go clean -modcache'
var cleanModCache bool

// lintAll stands for lint on all source code
var lintAll bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: "Go project boot",
	Long:  `Supply most frequently used tool and best practices for go project development`,
	ValidArgs: lo.Map(buildActions, func(item lo.Tuple2[string, common.ArgFunc], _ int) string {
		return item.A
	}),
	Args: cobra.MatchAll(cobra.OnlyValidArgs, cobra.MinimumNArgs(1)),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		//@todo validate the project according to gen.yml
		return nil
	},
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
	rootCmd.Flags().BoolVar(&cleanCache, CleanCacheFlag, false, "to remove the entire go build cache")
	rootCmd.Flags().BoolVar(&cleanTestCache, CleanTestCacheFlag, false, "to expire all test results in the go build cache")
	rootCmd.Flags().BoolVar(&cleanModCache, CleanModCacheFlag, false, "to remove the entire module download cache")
	rootCmd.Flags().BoolVar(&lintAll, LintAllFlag, false, "lint scan all source code, default only on changed source code")
}
