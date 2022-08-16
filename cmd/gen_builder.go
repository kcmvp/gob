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
	Short: "Generate build script for go currentProject",
	Long:  `Includes mostly used build actions: Clean, Test, Code Scan and Build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		project := currentProject(cmd)
		generateFile(builderTmp, filepath.Join(project.ScriptDir(), "builder.go"), nil)
		generateFile(golangCiTmp, ".golangci.yml", nil)
		return install(golangCi, "golangci-lint", "version")
	},
}

func init() {
	genCmd.AddCommand(builderCmd)
}
