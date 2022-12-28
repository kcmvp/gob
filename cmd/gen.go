/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"fmt"

	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/scaffolds"

	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:       string(boot.Generate),
	Short:     "Generate scaffolds & best practices. Run `gob gen -l` get full list",
	ValidArgs: scaffolds.ValidStack(string(boot.Generate)),
	Args: func(cmd *cobra.Command, args []string) error {
		if listCommandArgs {
			return nil
		}
		err := cobra.ExactArgs(1)(cmd, args)
		if err != nil {
			return fmt.Errorf("run with gen against current project:%w", err)
		}
		return cobra.OnlyValidArgs(cmd, args) //nolint
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		session := cmd.Context().Value(CurrentSession).(*boot.Session)
		session.BindFlag(boot.Generate, "stack", args[0])
		return session.Run(scaffolds.NewProject(cmd.Context().Value(RootDir).(string)), boot.Generate) //nolint
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}
