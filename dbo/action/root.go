/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package action

import (
	"embed"
	"errors"
	"github.com/kcmvp/app"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

const (
	dbxModule  = "github.com/kcmvp/dbo"
	baseEntity = "github.com/kcmvp/dbo/base"
	// AutoUpdateTime attribute tag with `aut` will be set to time.Now() for update
	AutoUpdateTime = "aut"
	// AutoCreateTime attribute tag with `act` will be set to time.Now() for creation
	AutoCreateTime = "act"
	// PK identify a column is primary key
	PK            = "pk"
	AttrSeparator = ";"
	//AttrIgnore        = "-"
	DefaultAttribute  = "default"
	DefaultStringSize = 10
	//TableFuncName     = "name"
)

//go:embed tmpl/*
var tempDir embed.FS

func validateRoot(cmd *cobra.Command, args []string) error {
	pws := mo.TupleToResult(os.Getwd())
	if pws.IsOk() {
		return lo.If(app.RootDir() != pws.MustGet(), errors.New("please run the command in project root")).ElseF(func() error {
			_, err := os.ReadFile(filepath.Join(app.RootDir(), "go.mod"))
			if err != nil {
				return errors.New("can not find go.mod in current directory")
			}
			return nil
		})
	}
	return pws.Error()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "dba",
	Short:             "got framework command line",
	Long:              `Use dba to generate dbx artifacts`,
	PersistentPreRunE: validateRoot,
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
