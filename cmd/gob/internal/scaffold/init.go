package scaffold

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/work"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: color.GreenString(`init scaffold for project`),
		Long:  color.GreenString(`init scaffold for project`),
		RunE: func(cmd *cobra.Command, args []string) error {
			module := work.Current()
			if module.IsError() {
				return module.Error()
			}
			fmt.Println("hello so nice to meet you!")
			return nil
		},
	}
}
