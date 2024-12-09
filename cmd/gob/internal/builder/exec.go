package builder

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"runtime"
	"strings"
)

func ExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec",
		Short: color.GreenString(`hook command`),
		Long:  color.GreenString(`it should be called from external event`),
		ValidArgs: func() []string {
			return lo.FilterMap(project.Plugins(), func(item project.Plugin, _ int) (string, bool) {
				// @todo need to review 2024-12-06 15:47:33
				return lo.Last(strings.Split(item.Name, "."))
			})
		}(),
		Args: func(cmd *cobra.Command, args []string) error {
			err := cobra.OnlyValidArgs(cmd, args)
			if err == nil {
				err = cobra.ExactArgs(1)(cmd, args)
			}
			return err
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var frame runtime.Frame
			more := true
			callers := make([]uintptr, 20)
			for {
				size := runtime.Callers(0, callers)
				if size == len(callers) {
					callers = make([]uintptr, 2*len(callers))
					continue
				}
				frames := runtime.CallersFrames(callers[:size])
				for more {
					frame, more = frames.Next()
					if strings.Contains(frame.File, args[0]) {
						return nil
					}
				}
				break
			}
			return fmt.Errorf("caller should be %s trigger", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			op := project.PluginByN(args[0])
			if op.IsAbsent() {
				return fmt.Errorf("can not find command %s", args[0])
			}
			return op.MustGet().Execute()
		},
	}
}
