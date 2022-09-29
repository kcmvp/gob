/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/kcmvp/gob/scaffolds"

	"github.com/kcmvp/gob/boot"
	"github.com/spf13/cobra"
)

var (
	version      string
	listAllSetup bool
)

const setupCommand = "setup"

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:       setupCommand,
	Short:     "Setup build script, hook, linter and other configuration. Run `gob setup -l` get more inforation",
	ValidArgs: scaffolds.ValidStack(setupCommand),
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.ExactArgs(1)(cmd, args)
		if err != nil && !listAllSetup {
			return fmt.Errorf("run with %s against current project:%w", setupCommand, err)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			session := cmd.Context().Value(CurrentSession).(*boot.Session)
			command := boot.Command(args[0])
			session.BindFlag(boot.InitLinter, "version", version)
			return session.Run(scaffolds.NewProject(), command) //nolint
		}
		scaffolds.ListStack(setupCommand)
		return nil
	},
}

func init() {
	setupCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
	setupCmd.Flags().BoolVarP(&listAllSetup, "list", "l", false, "list all available arguments")
	rootCmd.AddCommand(setupCmd)
}
