/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/kcmvp/gbt/builder"
	"golang.org/x/mod/modfile"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
)

var modules = []string{"github.com/kcmvp/gbt/builder"}

func importModule(ctx context.Context, module string, update bool) {
	f := ctx.Value(builder.MOD).(*modfile.File)
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
		if _, err := exec.Command("go", "get", module).CombinedOutput(); err != nil {
			fmt.Printf("failed to install module %s \n", module)
		}
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gbt",
	Short: "Generate go project scaffold",
	Long:  `Generate go project scaffolds (builder, hook)`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, module := range modules {
			importModule(cmd.Context(), module, false)
		}
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.ReadFile("go.mod")
		if err != nil {
			err = errors.New("please run the command in the module root directory")
		} else {
			ctx := context.WithValue(cmd.Context(), builder.MOD, f)
			cmd.SetContext(ctx)
		}
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gbt.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func generateFile(content string, targetName string, data interface{}) {
	dir := filepath.Dir(targetName)
	os.MkdirAll(dir, os.ModePerm)
	if f, err := os.OpenFile(targetName, os.O_RDWR|os.O_CREATE|os.O_EXCL, os.ModePerm); err == nil {
		defer f.Close()
		if t, err := template.New(targetName).Parse(content); err != nil {
			fmt.Println(fmt.Sprintf("Failed to parse template, %+v", err))
		} else {
			if err = t.Execute(f, data); err != nil {
				fmt.Println(fmt.Sprintf("Failed to create file %v, %+v", targetName, err))
			}
			abs, _ := filepath.Abs(f.Name())
			fmt.Println(fmt.Sprintf("generate files: %v successfully", abs))
		}
	} else {
		if errors.Is(err, os.ErrExist) {
			fmt.Printf("%s exists\n", targetName)
		} else {
			fmt.Printf("failed to generate file %s, %v\n", targetName, err)
		}
	}
}
