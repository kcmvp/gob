package scaffolds

import (
	_ "embed"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
	"github.com/thedevsaddam/gojsonq/v2"
)

type Stack struct {
	Name        string
	Command     string
	Module      string
	Description string
	DependsOn   string
	Register    bool
}

//go:embed template/stack.json
var stacks string

func searchStackByCategory(category string) interface{} {
	jq := gojsonq.New().FromString(stacks).From("stacks").Where("category", "=", category)
	return jq.Get()
}

func getStack(name string) Stack {
	data := gojsonq.New().FromString(stacks).From("stacks").Where("name", "=", name).Get()
	tm := data.([]interface{})[0].(map[string]interface{})
	stack := Stack{
		tm["name"].(string),
		"",
		tm["module"].(string),
		tm["Description"].(string),
		"",
		false,
	}
	if v, ok := tm["DependsOn"]; ok {
		stack.DependsOn = v.(string)
	}
	if v, ok := tm["Register"]; ok {
		stack.Register = v.(bool)
	}
	return stack
}

func ValidStack(category string) []string {
	data := searchStackByCategory(category)
	args := lo.Map(data.([]interface{}), func(t interface{}, i int) string {
		tm := t.(map[string]interface{})
		return tm["name"].(string)
	})
	return args
}

func ListStack(category string) []Stack {
	data := searchStackByCategory(category)
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
			tm["name"].(string),
			fmt.Sprintf("gob %s %s", category, tm["name"]),
			tm["module"].(string),
			tm["Description"].(string),
			"",
			false,
		}
		hasModule = hasModule || len(st.Module) == 0 || st.Module != "-"
		if v, ok := tm["DependsOn"]; ok {
			st.DependsOn = v.(string)
		}
		if v, ok := tm["Register"]; ok {
			st.Register = v.(bool)
		}
		return st
	})

	if len(validStacks) < 1 {
		return validStacks
	}

	if !hasModule {
		ct.AppendHeader(table.Row{"#", "Name", "Command", "Description"})
		ct.AppendRows(lo.Map(validStacks, func(t Stack, index int) table.Row {
			return table.Row{index + 1, t.Name, t.Command, t.Description}
		}))
	} else {
		ct.AppendHeader(table.Row{"#", "Name", "Module", "Command", "Description"})
		ct.AppendRows(lo.Map(validStacks, func(t Stack, index int) table.Row {
			return table.Row{index + 1, t.Name, t.Module, t.Command, t.Description}
		}))
	}
	fmt.Println(ct.Render())
	return validStacks
}
