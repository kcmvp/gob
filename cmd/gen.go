package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// genCmd represents the generate command
var genCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize useful infrastructures and tools",
	Long: `Initialize useful infrastructures and tools such as:
git hook, linter and so. run "gob init -h" get more information`,
	ValidArgsFunction: validArgsFun,
	Args: func(cmd *cobra.Command, args []string) error {
		return cobra.OnlyValidArgs(cmd, args) //nolint
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
	},
}

func init() {
	viper.SetDefault("ContentDir", "content")
}
