/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/kcmvp/gb/cmd/shared"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"strings"
)

const golangCiLinter = "github.com/golangci/golangci-lint/cmd/golangci-lint"
const defaultVersion = "v1.55.1"

var initFunc = func(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize configuration ......")
	_, ok := lo.Find(internal.CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string]) bool {
		return strings.HasPrefix(plugin.D, golangCiLinter)
	})
	if !ok {
		latest, err := shared.LatestVersion(golangCiLinter, "v1.55.*")
		if err != nil {
			latest = defaultVersion
		}
		internal.CurProject().InstallPlugin(fmt.Sprintf("%s@%s", golangCiLinter, latest), "lint")
	}
	internal.CurProject().SetupHook(true)
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project builder configuration",
	Long:  `Initialize project builder configuration`,
	Run:   initFunc,
}

func init() {
	rootCmd.AddCommand(initCmd)
}
