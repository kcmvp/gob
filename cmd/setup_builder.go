/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   boot.SetupBuilder.Name(),
	Short: "Setup build script for go current project",
	Long:  `Includes mostly used build actions: Clean, Test, Code scan and Build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		return session.Run(builder.NewBuilder(), boot.SetupBuilder) //nolint
	},
}

func init() {
	setupCmd.AddCommand(builderCmd)
}
