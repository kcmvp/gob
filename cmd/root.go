/*
Copyright © 2022 kcmvp <kcheng.mvp@gmail.com>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
	"github.com/kcmvp/gbt/builder"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

const (
	gbt            = "github.com/kcmvp/gbt"
	_ctxModFileKey = "mod"
	_ctxBuilder    = "builder"
)

var modules = []string{"github.com/kcmvp/gbt"}

func importModule(ctx context.Context, module string, update bool) {
	f := ctx.Value(_ctxModFileKey).(*modfile.File)
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
	Use:   "gbt",
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
			err = errors.New("please run the command in the module root directory")
		} else {
			if f, err := modfile.Parse("go.mod", data, nil); err != nil {
				return fmt.Errorf("invalid go.mod file")
			} else {
				ctx := context.WithValue(cmd.Context(), _ctxModFileKey, f)
				ctx = context.WithValue(ctx, _ctxBuilder, builder.NewBuilder(pwd))
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

func generateFile(content string, targetName string, data interface{}, trunk bool) {
	dir := filepath.Dir(targetName)
	os.MkdirAll(dir, os.ModePerm)
	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL
	if trunk {
		flag = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}
	if f, err := os.OpenFile(targetName, flag, os.ModePerm); err == nil {
		defer f.Close()
		if t, err := template.New(targetName).Parse(content); err != nil {
			log.Println(color.RedString("Failed to parse template, %+v", err))
		} else {
			if err = t.Execute(f, data); err != nil {
				log.Println(color.RedString("Failed to create file %v, %+v\n", targetName, err))
			} else {
				log.Printf("generate file %v successfully\n", f.Name())
			}
		}
	} else {
		if errors.Is(err, os.ErrExist) {
			log.Printf("%s exists\n", targetName)
		} else {
			log.Println(color.RedString("failed to generate file %s, %v\n", targetName, err))
		}
	}
}

func getBuilder(cmd *cobra.Command) *builder.Builder {
	if b, ok := cmd.Context().Value(_ctxBuilder).(*builder.Builder); ok {
		return b
	} else {
		log.Fatalln("failed to get current project")
	}
	return nil
}
