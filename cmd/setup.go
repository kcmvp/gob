package cmd

import (
	_ "embed"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kcmvp/gob/cmd/common"
	"github.com/kcmvp/gob/cmd/git"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/thedevsaddam/gojsonq/v2"
)

//go:embed setup.json
var data []byte

const (
	GitHook = "githook"
	Onion   = "onion"
	List    = "list"
)

type Setup struct {
	Name string
	Url  string
	Desc string
}

var setups []Setup
var list bool

var listFunc common.ArgFunc = func(cmd *cobra.Command) error {
	ct := table.Table{}
	ct.SetTitle("Available Setups")
	ct.AppendHeader(table.Row{"#", "Name", "Description"})
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	style.HTML.CSSClass = table.DefaultHTMLCSSClass
	ct.SetStyle(style)
	consoleRows := lo.Map(setups, func(setup Setup, i int) table.Row {
		return table.Row{i + 1, setup.Name, setup.Desc}
	})
	ct.AppendRows(consoleRows)
	fmt.Println(ct.Render())
	return nil
}

var setupFuncs = []lo.Tuple2[string, common.ArgFunc]{
	lo.T2(List, listFunc),
	lo.T2(Onion, onionFunc),
	lo.T2(GitHook, git.SetupGitHookFunc),
}

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup useful infrastructures and tools",
	Long: `Setup useful infrastructures and tools such as:
git hook, linter and so on. run "gob setup list" to list all supported artifacts`,
	ValidArgs: []string{List},
	Args:      cobra.MatchAll(cobra.OnlyValidArgs, cobra.MaximumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fn, _ := lo.Find(setupFuncs, func(item lo.Tuple2[string, common.ArgFunc]) bool {
				return args[0] == item.A
			})
			fn.B(cmd)
		}
	},
}

func init() {
	jq := gojsonq.New().FromString(string(data))
	v := jq.Select("name", "url", "desc").Get()
	for _, item := range v.([]interface{}) {
		m := item.(map[string]interface{})
		setups = append(setups, Setup{
			Name: fmt.Sprintf("%s", m["name"]),
			Url:  fmt.Sprintf("%s", m["url"]),
			Desc: fmt.Sprintf("%s", m["desc"]),
		})
		setupCmd.ValidArgs = append(setupCmd.ValidArgs, fmt.Sprintf("%s", m["name"]))
	}
	setupCmd.Flags().BoolVarP(&list, "list", "l", false, "list all artifactss")
}
