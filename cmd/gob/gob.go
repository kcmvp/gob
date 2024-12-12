/*
Copyright © 2023 kcheng.mvp@gmail.com
*/
package main

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/internal/scaffold"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: color.GreenString(`Go project boot`),
	Long: color.GreenString(`Go project boot, which include lots of useful builder plugins and scaffold
of frameworks`),
	//PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
	//	currentDir, _ := os.Getwd()
	//	fmt.Println(currentDir)
	//	fmt.Println(project.RootDir())
	//	if project.RootDir() != currentDir {
	//		return fmt.Errorf(color.RedString("Please execute the command in the project root dir"))
	//	}
	//	return nil
	//},
	//ValidArgs: func() []string {
	//	return lo.FilterMap(project.Plugins(), func(p project.Plugin, _ int) (string, bool) {
	//		return p.Name, len(p.Url) > 0
	//	})
	//}(),
	Args: func(cmd *cobra.Command, args []string) error {
		return cobra.MaximumNArgs(1)(cmd, args)
	},
	//RunE: func(cmd *cobra.Command, args []string) error {
	//	if len(args) > 0 {
	//		op := project.PluginByN(args[0])
	//		if op.IsAbsent() {
	//			return fmt.Errorf("can not find plugin %s", args[0])
	//		}
	//		return op.MustGet().Execute()
	//	}
	//	return nil
	//},
}

func init() {
	rootCmd.SetErrPrefix(color.RedString("Error:"))
	rootCmd.SetFlagErrorFunc(func(command *cobra.Command, err error) error {
		return lo.IfF(err != nil, func() error {
			return fmt.Errorf(color.RedString(err.Error()))
		}).Else(nil)
	})
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func main() {
	//rootCmd.AddCommand(builder.PluginCmd(), builder.DepCmd(), builder.ExecCmd(), scaffold.InitCmd(), scaffold.DboCmd())
	rootCmd.AddCommand(scaffold.InitCmd())
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
