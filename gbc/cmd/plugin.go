/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kcmvp/gob/gbc/artifact"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// alias is the tool alias
var alias string

// command is the tool command
var command string

// Install the specified tool as gob plugin
func install(_ *cobra.Command, args ...string) error {
	plugin, err := artifact.NewPlugin(args[0])
	if err != nil {
		return err
	}
	plugin.Alias = alias
	plugin.Alias = command
	artifact.CurProject().SetupPlugin(plugin)
	return nil
}

func list(_ *cobra.Command, _ ...string) error {
	plugins := artifact.CurProject().Plugins()
	ct := table.Table{}
	ct.SetTitle("Installed Plugins")
	ct.AppendRow(table.Row{"Command", "Plugin"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.Title.Align = text.AlignCenter
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	rows := lo.Map(plugins, func(plugin artifact.Plugin, index int) table.Row {
		return table.Row{plugin.Alias, plugin.Url}
	})
	ct.AppendRows(rows)
	fmt.Println(ct.Render())
	return nil
}

var pluginCmdAction = []Action{
	{
		A: "list",
		B: list,
		C: "list all setup plugins",
	},
	{
		A: "install",
		B: install,
		C: "install a plugin. `gbc plugin install <plugin url>`",
	},
}

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Install a new plugin or list installed plugins",
	Long: color.BlueString(`
Install a new plugin or list installed plugins
you can update the plugin by edit gob.yaml directly
`),
	Args: func(cmd *cobra.Command, args []string) error {
		if err := MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if !lo.Contains(lo.Map(pluginCmdAction, func(item Action, _ int) string {
			return item.A
		}), args[0]) {
			return errors.New(color.RedString("invalid argument %s", args[0]))
		}
		if "install" == args[0] && (len(args) < 2 || strings.TrimSpace(args[1]) == "") {
			return errors.New(color.RedString("miss the plugin url"))
		}
		return nil
	},
	ValidArgs: lo.Map(pluginCmdAction, func(item Action, _ int) string {
		return item.A
	}),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdAction, _ := lo.Find(pluginCmdAction, func(cmdAction Action) bool {
			return cmdAction.A == args[0]
		})
		return cmdAction.B(cmd, args[1:]...)
	},
}

func pluginExample() string {
	format := fmt.Sprintf("  %%-%ds %%s", pluginCmd.NamePadding())
	return strings.Join(lo.Map(pluginCmdAction, func(action Action, index_ int) string {
		return fmt.Sprintf(format, action.A, action.C)
	}), "\n")
}

func init() {
	// init pluginCmd
	pluginCmd.Example = pluginExample()
	pluginCmd.SetUsageTemplate(usageTemplate())
	pluginCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	pluginCmd.Flags().StringVarP(&alias, "alias", "a", "", "alias of the tool")
	pluginCmd.Flags().StringVarP(&command, "command", "c", "", "default command of this tool")

	rootCmd.AddCommand(pluginCmd)
}
