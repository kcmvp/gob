package scaffold

import (
	_ "embed"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/buildtime"
	"github.com/olekukonko/tablewriter"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
	"os"
)

var (
	//go:embed resources/modules.json
	modules   string
	validArgs []string
)

func init() {
	rs := gjson.Get(modules, "modules")
	validArgs = lo.FilterMap(rs.Array(), func(rs gjson.Result, _ int) (string, bool) {
		return rs.Map()["name"].String(), !(rs.Map()["builtin"].Bool() && !rs.Map()["command"].Bool())
	})
}
func helpF(command *cobra.Command, strings []string) {
	rs := gjson.Get(modules, "modules")
	rows := lo.FilterMap(rs.Array(), func(rs gjson.Result, _ int) ([]string, bool) {
		tmp := rs.Map()
		return []string{tmp["name"].String(), tmp["module"].String(), tmp["category"].String(), lo.If(tmp["builtin"].Bool(), "Y").Else("N")}, true
	})
	fmt.Println(rows)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Module", "BuiltIn"})
	table.SetFooter([]string{"", "Total", "$146.93"}) // Add Footer
}

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:       "init",
		Short:     color.GreenString(`init scaffold for project`),
		Long:      color.GreenString(`init scaffold for project`),
		ValidArgs: validArgs,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.OnlyValidArgs(cmd, args); err != nil {
				return err
			}
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			module := buildtime.Current()
			if module.IsError() {
				return module.Error()
			}
			return nil
		},
	}
	initCmd.SetHelpFunc(helpF)
	return initCmd
}
