/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"github.com/kcmvp/gob/cmd/action"
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

//go:embed resources/version.tmpl
var versionTmp []byte

func setupVersion() {
	// check config folder
	// create version.go
	infra := filepath.Join(internal.CurProject().Root(), "infra")
	if _, err := os.Stat(infra); err != nil {
		os.Mkdir(infra, os.ModePerm) // nolint
	}
	ver := filepath.Join(infra, "version.go")
	if _, err := os.Stat(ver); err != nil {
		os.WriteFile(ver, versionTmp, 0o666) //nolint
	}
}

var initializerFunc = func(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize configuration ......")
	_, ok := lo.Find(internal.CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string]) bool {
		return strings.HasPrefix(plugin.D, golangCiLinter)
	})
	if !ok {
		latest, err := action.LatestVersion(golangCiLinter, "v1.55.*")
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
	setupVersion()
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
