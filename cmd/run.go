/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
)

var scanAll = false

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:       "run",
	Short:     "run clean, test, lint and build against current project",
	ValidArgs: []string{"clean", "test", "lint", "build"},
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.MinimumNArgs(1)(cmd, args)
		if err == nil {
			err = cobra.OnlyValidArgs(cmd, args)
		}
		return err
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%v \n", args)
		var acts []builder.Action
		for _, act := range args {
			if a, ok := builder.RunAction(act); ok {
				// @todo review the design
				if len(acts) > 0 && a == acts[len(acts)-1] {
					log.Printf("ignore repeat action %s \n", a)
				} else {
					acts = append(acts, a)
				}
			}
		}
		ctx := context.WithValue(cmd.Context(), builder.ScanAll, scanAll) //nolint
		getBuilder(ctx).RunCtx(ctx, acts...)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&scanAll, "scan-all", "a", false, "scan all the source code")
}
