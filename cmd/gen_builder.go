/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"os"


	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Generate build script for go current project",
	Long:  `Includes mostly used build actions: Clean, Test, Code scan and Build`,
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := os.Getwd()
		builder.NewBuilder(root).Run(cmd.Name())
	},
}

func init() {
	genCmd.AddCommand(builderCmd)
}
