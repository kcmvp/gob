/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kcmvp/gob/cmd/shared"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

const golangCiLinter = "github.com/golangci/golangci-lint/cmd/golangci-lint"
const defaultVersion = "v1.55.1"

//go:embed resources/.golangci.yaml
var golangci []byte

var initializerFunc = func(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize configuration ......")
	_, ok := lo.Find(internal.CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string]) bool {
		return strings.HasPrefix(plugin.D, golangCiLinter)
	})
	if !ok {
		latest, err := shared.LatestVersion(golangCiLinter, "v1.55.*")
		if err != nil {
			latest = defaultVersion
		}
		internal.CurProject().InstallPlugin(fmt.Sprintf("%s@%s", golangCiLinter, latest), "lint", "run, ./...")
		cfg := filepath.Join(internal.CurProject().Root(), ".golangci.yaml")
		if _, err := os.Stat(cfg); err != nil {
			os.WriteFile(cfg, golangci, 0666)
		}
	}
	internal.CurProject().Setup(true)
}

// initializerCmd represents the init command
var initializerCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project builder configuration",
	Long:  `Initialize project builder configuration`,
	Run:   initializerFunc,
}

func init() {
	builderCmd.AddCommand(initializerCmd)
}
