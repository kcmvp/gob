/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/kcmvp/gob/scaffolds"

	"github.com/kcmvp/gob/boot"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ContextKey string

var CurrentSession ContextKey = "session"

var listCommandArgs bool

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: "Golang project boot",
	Long: `
Generate project scaffolds, including build script, github hook and other project setup.
Please visit https://github.com/kcmvp/gob/wiki for details`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if listCommandArgs {
			sts := scaffolds.ListStack(cmd.Name())
			if len(sts) > 0 {
				os.Exit(0)
			}
		}
		_, err := os.ReadFile("go.mod")
		if err != nil {
			err = errors.New(color.RedString("please run the command in the module root directory"))
		}
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ctx := context.WithValue(context.Background(), CurrentSession, boot.NewSession())
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&listCommandArgs, "list", "l", false, "list all available arguments for command")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
