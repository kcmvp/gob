/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"github.com/kcmvp/gob/scaffolds"

	"github.com/kcmvp/gob/boot"
	"github.com/spf13/cobra"
)

var version string

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:       "setup",
	Short:     "Setup build script, hook, linter and other configurations. Run `gob setup -l` get more information",
	ValidArgs: scaffolds.ValidStack("setup"),
	Args: func(cmd *cobra.Command, args []string) error {
		if listCommandArgs {
			return nil
		}
		err := cobra.ExactArgs(1)(cmd, args)
		if err != nil {
			return err
		}
		return cobra.OnlyValidArgs(cmd, args) //nolint
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		command := boot.Command(args[0])
		session.BindFlag(boot.SetupLinter, "version", version)
		return session.Run(scaffolds.NewProject(), command) //nolint
	},
}

func init() {
	setupCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
	rootCmd.AddCommand(setupCmd)
}
