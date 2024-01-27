/*
Copyright © 2023 kcmvp <kcheng.mvp@gmail.com>
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
	var data []byte
	var err error
	test, _ := internal.TestCaller()
	if !test {
		data, err = resources.ReadFile("resources/config.json")
	} else {
		data, err = os.ReadFile(filepath.Join(internal.CurProject().Root(), "testdata", "config.json"))
	}
	var plugins []internal.Plugin
	if err == nil {
		v := gjson.GetBytes(data, "plugins")
		if err = json.Unmarshal([]byte(v.Raw), &plugins); err != nil {
			color.Red("failed to parse plugin %s", err.Error())
		}
	}
	return plugins
}

func initializerFunc(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize configuration ......")
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
