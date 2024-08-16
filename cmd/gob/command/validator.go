package command

import (
	"errors"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func MinimumNArgs(n int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) < n {
			return errors.New(color.RedString("requires at least %d arg(s), only received %d", n, len(args)))
		}
		return nil
	}
}

func OnlyValidArgs(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if !lo.Contains(cmd.ValidArgs, arg) {
			return errors.New(color.RedString("invalid argument %q for %s", arg, cmd.CommandPath()))
		}
	}
	return nil
}
