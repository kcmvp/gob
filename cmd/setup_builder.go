/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
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
		return boot.Run(builder.NewBuilder(), boot.SetupBuilder)
	},
}

func init() {
	setupCmd.AddCommand(builderCmd)
}
