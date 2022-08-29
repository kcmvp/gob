/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Generate build script for go current project",
	Long:  `Includes mostly used build actions: Clean, Test, Code Scan and Build`,
	Run: func(cmd *cobra.Command, args []string) {
		getBuilder(cmd.Context()).Run(builder.SetupBuilder)
	},
}

func init() {
	genCmd.AddCommand(builderCmd)
}
