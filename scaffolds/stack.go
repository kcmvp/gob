package scaffolds

import (
	_ "embed"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
	"github.com/thedevsaddam/gojsonq/v2"
)

type Stack struct {
	Index       int
	Name        string
	Command     string
	Module      string
	Description string
}

//go:embed template/stack.json
var stacks string

func byCategory(category string) interface{} {
	jq := gojsonq.New().FromString(stacks).From("stacks").Where("category", "=", category)
	return jq.Get()
}

func ValidStack(category string) []string {
	data := byCategory(category)
	args := lo.Map(data.([]interface{}), func(t interface{}, i int) string {
		tm := t.(map[string]interface{})
		return tm["name"].(string)
	})
	return args
}

func ListStack(category string) []Stack {
	data := byCategory(category)
	ct := table.Table{}
	ct.SetTitle(fmt.Sprintf("All valid arguments for `%s`, you can run corresponding command to take the action", category))
	style := table.StyleDefault
	style.Options.DrawBorder = true
	style.Options.SeparateRows = true
	style.Options.SeparateColumns = true
	ct.SetStyle(style)
	hasModule := false
	validStacks := lo.Map(data.([]interface{}), func(t interface{}, i int) Stack {
		tm := t.(map[string]interface{})
		st := Stack{
			i + 1,
			tm["name"].(string),
			fmt.Sprintf("gob %s %s", category, tm["name"]),
			tm["module"].(string),
			tm["Description"].(string),
		}
		hasModule = hasModule || len(st.Module) == 0 || st.Module != "-"
		return st
	})

	if !hasModule {
		ct.AppendHeader(table.Row{"#", "Name", "Command", "Description"})
		ct.AppendRows(lo.Map(validStacks, func(t Stack, _ int) table.Row {
			return table.Row{t.Index, t.Name, t.Command, t.Description}
		}))
	} else {
		ct.AppendHeader(table.Row{"#", "Name", "Module", "Command", "Description"})
		ct.AppendRows(lo.Map(validStacks, func(t Stack, _ int) table.Row {
			return table.Row{t.Index, t.Name, t.Module, t.Command, t.Description}
		}))
	}
	fmt.Println(ct.Render())
	return validStacks
}
