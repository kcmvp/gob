package plugin

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/kcmvp/gb/cmd/action"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

// ToolAlias is the tool alias, for the convenience of run 'gb alias'
var ToolAlias string

// ToolCommand is the tool command, it's the default command when run 'gb alias'
var ToolCommand string

// Install the specified tool as gb plugin
var Install action.Execution = func(cmd *cobra.Command, args ...string) error {
	if strings.HasSuffix(args[0], "@master") || strings.HasSuffix(args[0], "@latest") {
		return fmt.Errorf("please use specific version instead of 'master' or 'latest'")
	}
	err := internal.CurProject().InstallPlugin(args[0], ToolAlias, ToolCommand)
	if errors.Is(err, internal.PluginExists) {
		color.Yellow("Plugin %s exists", args[0])
		err = nil
	}
	return err
}

var UpdateList bool
var List action.Execution = func(cmd *cobra.Command, _ ...string) error {
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
	action.PrintCmd(cmd, ct.Render())
	if UpdateList {
		for _, plugin := range plugins {
			_, name := internal.NormalizePlugin(plugin.D)
			if _, err := os.Stat(filepath.Join(os.Getenv("GOPATH"), "bin", name)); err != nil {
				if err = internal.CurProject().InstallPlugin(plugin.D, plugin.A, plugin.C); err != nil {
					color.Yellow("Waring: failed to install %s", plugin.D)
				}
			} else {
				fmt.Printf("%s exists on the system \n", plugin.D)
			}
		}
	}
	return nil
}
