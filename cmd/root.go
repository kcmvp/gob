/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"os"
)

const (
	gbt           = "github.com/kcmvp/gbt"
	ctxModFileKey = "mod"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: "Generate project scaffold",
	Long:  `Generate project scaffolds (builder, hook)`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile("go.mod")
		if err != nil {
			err = errors.New(color.RedString("please run the command in the module root directory"))
		} else {
			if f, err := modfile.Parse("go.mod", data, nil); err != nil {
				return fmt.Errorf("invalid go.mod file")
			} else {
				ctx := context.WithValue(cmd.Context(), ctxModFileKey, f) //nolint
				cmd.SetContext(ctx)
			}
		}
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
