package cmd

import (
	"fmt"

	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

type (
	Execution func(cmd *cobra.Command, args ...string) error
	Action    lo.Tuple2[string, Execution]
)

func execute(cmd *cobra.Command, arg string) error {
	msg := fmt.Sprintf("Start %s project", arg)
	fmt.Printf("%-20s ...... \n", msg)
	if plugin, ok := lo.Find(internal.CurProject().Plugins(), func(plugin internal.Plugin) bool {
		return plugin.Alias == arg
	}); ok {
		return plugin.Execute()
	} else if action, ok := lo.Find(builtinActions, func(action Action) bool {
		return action.A == arg
	}); ok {
		return action.B(cmd, arg)
	}
	return fmt.Errorf("can not find command %s", arg)
}
