/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
)

func builtinPlugins() []internal.Plugin {
	data, err := resources.ReadFile("resources/config.json")
	var plugins []internal.Plugin
	if err == nil {
		v := gjson.GetBytes(data, "plugins")
		if err = json.Unmarshal([]byte(v.Raw), &plugins); err != nil {
			color.Red("failed to parse plugin %s", err.Error())
		}
	}
	return plugins
}

func initBuildVersion() {
	infra := filepath.Join(internal.CurProject().Root(), "infra")
	if _, err := os.Stat(infra); err != nil {
		os.Mkdir(infra, 0700) // nolint
	}
	ver := filepath.Join(infra, "version.go")
	if _, err := os.Stat(ver); err != nil {
		data, _ := resources.ReadFile(filepath.Join(resourceDir, "version.tmpl"))
		os.WriteFile(ver, data, 0666) //nolint
	}
}

func initializerFunc(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize configuration ......")
	initBuildVersion()
	lo.ForEach(builtinPlugins(), func(plugin internal.Plugin, index int) {
		internal.CurProject().SetupPlugin(plugin)
		if len(plugin.Config) > 0 {
			if _, err := os.Stat(filepath.Join(internal.CurProject().Root(), plugin.Config)); err != nil {
				if data, err := resources.ReadFile(filepath.Join(resourceDir, plugin.Config)); err == nil {
					if err = os.WriteFile(filepath.Join(internal.CurProject().Root(), plugin.Config), data, os.ModePerm); err != nil {
						color.Red("failed to create configuration %s", plugin.Config)
					}
				} else {
					color.Red("can not find the configuration %s", plugin.Config)
				}
			}
		}
	})
	internal.CurProject().SetupHooks(true)
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
