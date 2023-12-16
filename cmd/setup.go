package cmd

import (
	_ "embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/action"
	"github.com/kcmvp/gob/cmd/setup"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"strings"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup useful infrastructures and tools",
	Long: `Setup useful infrastructures and tools 
Run 'gob setup list get full supported list'`,
	ValidArgs: func() []string {
		return lo.Map(setup.Actions, func(item action.CmdAction, index int) string {
			return item.A
		})
	}(),
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MatchAll(cobra.OnlyValidArgs, cobra.ExactArgs(1))(cmd, args); err != nil {
			return fmt.Errorf(color.RedString(err.Error()))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fn, _ := lo.Find(setup.Actions, func(item action.CmdAction) bool {
			return item.A == args[0]
		})
		return fn.B(cmd, args...)
	},
}

func init() {
	setupCmd.SetUsageFunc(func(command *cobra.Command) error {
		keys := lo.Map(setup.Actions, func(item action.CmdAction, _ int) string {
			return item.A
		})
		color.HiYellow("Valid Arguments: %s", strings.Join(keys, ", "))
		return nil
	})
	rootCmd.AddCommand(setupCmd)
}
