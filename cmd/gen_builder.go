/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	_ "embed"
	"log"
	"path/filepath"

	"github.com/kcmvp/gos/infra"
	"github.com/spf13/cobra"
)

//go:embed template/builder.tmpl
var builderTmp string

// builderCmd represents the builder command.
var builderCmd = &cobra.Command{
	Use:   "builder",
	Short: "Generate build script for go current project",
	Long:  `Includes mostly used build actions: Clean, Test, Code Scan and Build`,
	RunE: func(cmd *cobra.Command, args []string) error {
		builder := getBuilder(cmd)
		log.Println("generating `builder.go`")
		return infra.GenerateFile(builderTmp, filepath.Join(builder.ScriptDir(), "builder.go"), nil, false) //nolint:wrapcheck
	},
}

func init() {
	setupCmd.AddCommand(builderCmd)
}
