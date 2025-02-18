package action

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:       "init",
	Short:     "init scaffold for project",
	Long:      "init scaffold for project",
	ValidArgs: []string{"mysql", "pg"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
}

func init() {
	rootCmd.AddCommand(initCmd)
}
