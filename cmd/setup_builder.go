/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"github.com/kcmvp/gob/boot"
	"os"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Setup build script for go current project",
	Long:  `Includes mostly used build actions: Clean, Test, Code scan and Build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _ := os.Getwd()
		executor := boot.NewExecutor()
		return executor.Run(builder.NewBuilder(root), cmd.Name())
	},
}

func init() {
	setupCmd.AddCommand(builderCmd)
}
