/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"github.com/kcmvp/gbt/builder"
	"github.com/spf13/cobra"
	"path/filepath"
)

//go:embed template/builder.tmpl
var builderTmp string

// builderCmd represents the builder command
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Generate build script for go project",
	Long:  `Includes mostly used build actions: Clean, Test, Code Scan and Build`,
	Run: func(cmd *cobra.Command, args []string) {
		generateFile(builderTmp, filepath.Join(builder.ScriptsDir, "builder.go"), nil)
		builder.GolangCiLinter.Install()
		generateFile(builder.GolangCiLinter.Content(), builder.GolangCiLinter.ConfigName(), nil)
	},
}

func init() {
	rootCmd.AddCommand(builderCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// builderCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// builderCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
