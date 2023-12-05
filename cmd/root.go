// Package cmd /*
package cmd

import (
	"context"
	"github.com/kcmvp/gob/internal"
	"github.com/spf13/cobra"
	"os"
)

var validArgsFun = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var validArgs []string
	if cmd.Name() == "gob" {
		validArgs = append(validArgs, []string{"build", "clean", "test", "package"}...)
	}
	return validArgs, cobra.ShellCompDirectiveNoSpace
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "gob",
	Short:             "Go project boot",
	Long:              `Supply most frequently usage and best practice for go project development`,
	ValidArgsFunction: validArgsFun,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gob.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.SetErrPrefix(internal.Red.Sprintf("Error:"))
	rootCmd.AddCommand(initCmd)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
