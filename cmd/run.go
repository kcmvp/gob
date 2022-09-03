/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       "run",
	Short:     fmt.Sprintf("Valid run flags are: %s", strings.Join(builder.Children("run"), ",")),
	ValidArgs: builder.Children("run"),
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Executing commands :%s\n", strings.Join(args, ","))
		flags := []string{}
		ctx := context.WithValue(cmd.Context(), builder.CtxKeyRunFlags, flags)
		builder.RunCtx(ctx, args...)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
