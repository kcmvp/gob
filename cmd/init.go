/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/kcmvp/gob/scaffolds"

	"github.com/kcmvp/gob/boot"
	"github.com/spf13/cobra"
)

var version string

// initCmd represents the setup command.
var initCmd = &cobra.Command{
	Use:       string(boot.Init),
	Short:     strings.Join(boot.ValidCommands(boot.Init), ","),
	ValidArgs: boot.ValidCommands(boot.Init),
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.ExactArgs(1)(cmd, args)
		if err != nil {
			err = fmt.Errorf("%w. Run with: %s against current project", err, strings.Join(boot.ValidCommands(boot.Init), ","))
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		command := boot.Command(args[0])
		session.BindFlag(boot.InitLinter, "version", version)
		return session.Run(scaffolds.NewProject(), command) //nolint
	},
}

func init() {
	initCmd.Flags().StringVarP(&version, "version", "v", boot.LatestVer, "golangci-lint version")
	rootCmd.AddCommand(initCmd)
}
