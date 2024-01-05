/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kcmvp/gob/cmd/action"
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
func install(_ *cobra.Command, args ...string) error {
	plugin, err := internal.NewPlugin(args[0])
	if err != nil {
		return err
	}
	internal.CurProject().SetupPlugin(plugin)
	return nil
}

func list(_ *cobra.Command, _ ...string) error {
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
	rows := lo.Map(plugins, func(plugin internal.Plugin, index int) table.Row {
		return table.Row{plugin.Name(), plugin.Command, plugin.Args, plugin.Url}
	})
	ct.AppendRows(rows)
	fmt.Println(ct.Render())
	return nil
}

var pluginCmdAction = []action.CmdAction{
	{
		A: "list",
		B: list,
	},
	{
		A: "install",
		B: install,
	},
}

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Install a new plugin or list installed plugins",
	Long: `Install a new plugin or list installed plugins
you can update the plugin by edit gob.yaml directly
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if !lo.Contains(lo.Map(pluginCmdAction, func(item action.CmdAction, _ int) string {
			return item.A
		}), args[0]) {
			return fmt.Errorf("invalid argument %s", args[0])
		}
		if "install" == args[0] && (len(args) < 2 || strings.TrimSpace(args[1]) == "") {
			return errors.New("miss the plugin url")
		}
		return nil
	},
	ValidArgs: lo.Map(pluginCmdAction, func(item action.CmdAction, _ int) string {
		return item.A
	}),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdAction, _ := lo.Find(pluginCmdAction, func(cmdAction action.CmdAction) bool {
			return cmdAction.A == args[0]
		})
		return cmdAction.B(cmd, args[1:]...)
	},
}

func init() {
	// init pluginCmd
	builderCmd.AddCommand(pluginCmd)
	// init installPluginCmd
	pluginCmd.Flags().StringVarP(&alias, "alias", "a", "", "alias of the tool")
	pluginCmd.Flags().StringVarP(&command, "command", "c", "", "default command of this tool")
}
