/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"path/filepath"

	"github.com/spf13/cobra"
)

const golangCi = "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

//go:embed template/.golangci.yml
var golangCiTmp string

//go:embed template/builder.tmpl
var builderTmp string

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Generate build script for go project",
	Long:  `Includes mostly used build actions: Clean, Test, Code Scan and Build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		generateFile(cmd.Context(), builderTmp, filepath.Join(scriptDir, "builder.go"), nil)
		generateFile(cmd.Context(), golangCiTmp, ".golangci.yml", nil)
		return install(golangCi, "golangci-lint", "version")
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
