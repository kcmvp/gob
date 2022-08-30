/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"

	"github.com/kcmvp/gob/builder"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

const (
	gbt           = "github.com/kcmvp/gbt"
	ctxModFileKey = "mod"
)

var modules = []string{"github.com/kcmvp/gbt"}

func importModule(ctx context.Context, module string, update bool) {
	f := ctx.Value(ctxModFileKey).(*modfile.File)
	if strings.EqualFold(gbt, f.Module.Mod.Path) {
		return
	}
	has := false
	for _, require := range f.Require {
		if has = require.Mod.Path == module; has {
			break
		}
	}
	if !has || update {
		action := "installing"
		if has {
			action = "updating"
		}
		fmt.Printf("%s %s \n", action, module)
		if out, err := exec.Command("go", "get", module).CombinedOutput(); err != nil {
			fmt.Printf("failed to install module %s \n", module)
		} else {
			fmt.Println(string(out))
		}
	}
}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "gob",
	Short: "Generate project scaffold",
	Long:  `Generate project scaffolds (builder, hook)`,
	// Run: func(cmd *cobra.Command, args []string) {
	//	for _, module := range modules {
	//		importModule(cmd.Context(), module, false)
	//	}
	// },
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		pwd, _ := os.Getwd()
		data, err := os.ReadFile("go.mod")
		if err != nil {
			err = errors.New(color.RedString("please run the command in the module root directory"))
		} else {
			if f, err := modfile.Parse("go.mod", data, nil); err != nil {
				return fmt.Errorf("invalid go.mod file")
			} else {
				ctx := context.WithValue(cmd.Context(), ctxModFileKey, f) //nolint
				ctx = context.WithValue(ctx, builder.CtxKeyBuilder, builder.NewBuilder(pwd))
				cmd.SetContext(ctx)
			}
		}
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
