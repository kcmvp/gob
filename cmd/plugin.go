/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	//nolint
	"strings"
)

// alias is the tool alias
var alias string

// command is the tool command
var command string

// Install the specified tool as gob plugin
func install(args ...string) error {
	if strings.HasSuffix(args[0], "@master") || strings.HasSuffix(args[0], "@latest") {
		return fmt.Errorf("please use specific version instead of 'master' or 'latest'")
	}
	err := internal.CurProject().InstallPlugin(args[0], alias, command)
	if errors.Is(err, internal.PluginExists) {
		color.Yellow("Plugin %s exists", args[0])
		err = nil
	}
	return err
}

func list() {
	plugins := internal.CurProject().Plugins()
	ct := table.Table{}
	ct.SetTitle("Installed Plugins")
	ct.AppendRow(table.Row{"Command", "Alias", "Method", "URL"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.Title.Align = text.AlignCenter
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	rows := lo.Map(plugins, func(item lo.Tuple4[string, string, string, string], index int) table.Row {
		return table.Row{item.A, item.B, item.C, item.D}
	})
	ct.AppendRows(rows)
	fmt.Println(ct.Render())
}

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "List all configured plugins",
	Long:  `List all configured plugins`,
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

// installPluginCmd represents the plugin install command
var installPluginCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a tool as gob plugin",
	Long:  `Install a tool as gob plugin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf(color.RedString(err.Error()))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return install(args...)
	},
}

func init() {
	// init pluginCmd
	builderCmd.AddCommand(pluginCmd)

	// init installPluginCmd
	pluginCmd.AddCommand(installPluginCmd)
	installPluginCmd.Flags().StringVarP(&alias, "alias", "a", "", "alias of the tool")
	installPluginCmd.Flags().StringVarP(&command, "command", "c", "", "default command of this tool")
}
