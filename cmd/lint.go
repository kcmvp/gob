/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/kcmvp/gbt/builder"
	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "run code lint against project",
	Run: func(cmd *cobra.Command, args []string) {
		getBuilder(cmd).Run(builder.Lint)
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lintCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lintCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
