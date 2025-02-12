package scaffold

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func DboCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dbo",
		Short: color.GreenString(`Generate schema or repository for entity`),
		Long:  color.GreenString(`Generate schema or repository for entity`),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			//currentDir, _ := os.Getwd()
			//if project.RootDir() != currentDir {
			//	return fmt.Errorf(color.RedString("Please execute the command in the project root dir"))
			//}
			//return nil
			return nil
		},
	}
}
