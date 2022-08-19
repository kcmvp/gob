/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kcmvp/gos/builder"
	"github.com/spf13/cobra"
)

// testCmd represents the test command.
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "run unit test",
	Run: func(cmd *cobra.Command, args []string) {
		getBuilder(cmd).Run(builder.Test)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
