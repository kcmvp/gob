package scaffold

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/spf13/cobra"
	"os"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: color.GreenString(`init scaffold for project`),
		Long:  color.GreenString(`init scaffold for project`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			currentDir, _ := os.Getwd()
			if project.RootDir() != currentDir {
				return fmt.Errorf(color.RedString("Please execute the command in the project root dir"))
			}
			return nil
		},
	}
}
