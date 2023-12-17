/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gb/cmd/plugin"
	"github.com/spf13/cobra"
)

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "List all configured plugins",
	Long:  `List all configured plugins`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// run 'gb plugin' will list all the configured plugins
		// run 'gb plugin -u' will list all the configured plugins and install the uninstalled tools.
		return plugin.List(cmd)
	},
}

// installPluginCmd represents the plugin install command
var installPluginCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a tool as gb plugin",
	Long:  `Install a tool as gb plugin`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf(color.RedString(err.Error()))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return plugin.Install(cmd, args...)
	},
}

func init() {
	// init pluginCmd
	rootCmd.AddCommand(pluginCmd)
	pluginCmd.Flags().BoolVarP(&plugin.UpdateList, "update", "u", false, "update configured plugins")

	// init installPluginCmd
	pluginCmd.AddCommand(installPluginCmd)
	installPluginCmd.Flags().StringVarP(&plugin.ToolAlias, "alias", "a", "", "alias of the tool")
	installPluginCmd.Flags().StringVarP(&plugin.ToolCommand, "command", "c", "", "default command of this tool")
}
