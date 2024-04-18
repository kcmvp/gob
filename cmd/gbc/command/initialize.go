/*
Copyright © 2023 kcmvp <kcheng.mvp@gmail.com>
*/
package command

import (
	"fmt"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/spf13/cobra"
)

func initProject(_ *cobra.Command, _ []string) {
	fmt.Println("Initialize project ......")
	artifact.CurProject().SetupHooks(true)
}

// initializerCmd represents the init command
var initializerCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project",
	Long: `It will generate main.go file together with below initialization:
	1: Init git hooks
	2: Setup plugin(test,lint)
	3: Initialize application.yaml
	4: Necessary dependencies`,
	Run: initProject,
}

func init() {
	rootCmd.AddCommand(initializerCmd)
}
