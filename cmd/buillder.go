package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	CleanCmd = "clean"
	TestCmd  = "test"
	LintCmd  = "lint"
	BuildCmd = "build"
)

var blueMsg = color.New(color.FgCyan)

var builderFuncs = []lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]{
	lo.T2(CleanCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(CleanCmd)
		return nil
	}),
	lo.T2(TestCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(TestCmd)
		return nil
	}),
	lo.T2(LintCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(LintCmd)
		return nil
	}),
	lo.T2(BuildCmd, func(cmd *cobra.Command, args []string) error {
		fmt.Println(BuildCmd)
		return nil
	}),
}

var build = func(cmd *cobra.Command, args []string) error {
	var err error
	if LintCmd == cmd.Name() {
		if t, ok := lo.Find(builderFuncs, func(item lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]) bool {
			return LintCmd == item.A
		}); ok {
			fmt.Println(blueMsg.Sprintf("Start %s", cmd.Name()))
			err = t.B(cmd, args)
		}
	} else {
		fns := lo.DropRightWhile(builderFuncs, func(t2 lo.Tuple2[string, func(cmd *cobra.Command, args []string) error]) bool {
			return t2.A != cmd.Name()
		})
		for _, fn := range fns {
			fmt.Println(blueMsg.Sprintf("Start %s", fn.A))
			if err = fn.B(cmd, args); err != nil {
				return err
			}
		}
	}
	return err
}

var cleanCmd = &cobra.Command{
	Use:   CleanCmd,
	Short: "Clean target folder and build caches",
	Long:  `Clean target folder and build caches`,
	RunE:  build,
}

var testCmd = &cobra.Command{
	Use:   TestCmd,
	Short: "Run tests of the project",
	Long:  `Run tests of the project and test report will be generated at ${root}/target`,
	RunE:  build,
}

var lintCmd = &cobra.Command{
	Use:   LintCmd,
	Short: "Build all main packages in project",
	Long:  `Build all main packages in project`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// a warning message will be print when golangci-lint is not found and lintCmd is called separately
		// and gob will try to install golangci-lint
		// @todo
		// a warning message will be print when golangci-lint is not found and lintCmd is called separately
		// @todo
		return nil
	},
	RunE: build,
}
var buildCmd = &cobra.Command{
	Use:   BuildCmd,
	Short: "Build all main packages in the project",
	Long:  `Build all main packages in the project`,
	RunE:  build,
}

func init() {
	viper.SetDefault("ContentDir", "content")
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(buildCmd)
}
