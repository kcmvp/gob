/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/scaffolds"

	"github.com/spf13/cobra"
)

// showCmd represents the show command.
var showCmd = &cobra.Command{
	Use:   string(boot.InitList),
	Short: "Show all support scaffolds",
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		return session.Run(scaffolds.NewProject(), boot.InitList) //nolint
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
