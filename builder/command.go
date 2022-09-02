package builder

import (
	"context"
	//nolint
	_ "embed"
	"encoding/json"
	"log"
	"strings"

	"github.com/fatih/color"
)

//go:embed commands.json
var commandsJSON string

type Command struct {
	Parent   string
	Name     string
	Output   []string
	Internal bool
	Flags    []string
	stack    []string
}

func commandMap() map[string]Command {
	var commands []Command
	err := json.Unmarshal([]byte(commandsJSON), &commands)
	if err != nil {
		log.Fatalf("failed to read command json:%s\n", err.Error())
	}
	cm := map[string]Command{}
	for _, command := range commands {
		cm[command.Name] = command
	}
	return cm
}

func Actions(cmdName string) []Action {
	acm := map[string][]Action{
		"pre_commit.go": {cleanFunc, testFunc, lintFunc},
		"commit_msg.go": {commitMsgFunc},
		"pre_push.go":   {cleanFunc, testFunc},
		"builder":       {createDirFunc, builderFunc},
		"gitHook":       {createDirFunc, gitHookFunc},
		"clean":         {cleanFunc, gitHookFunc},
		"lint":          {createDirFunc, gitHookFunc, lintFunc},
		"test":          {createDirFunc, gitHookFunc, testFunc},
		"build":         {createDirFunc, gitHookFunc, buildFunc},
	}
	return acm[cmdName]
}

func Children(parent string) []string {
	var children []string
	for _, command := range commandMap() {
		if command.Parent == parent {
			children = append(children, command.Name)
		}
	}
	return children
}

func processCommands(ctx context.Context, cmds ...string) {
	cm := commandMap()
	var err error
	for _, cmd := range cmds {
		if c, ok := cm[cmd]; ok {
			if ctx, err = c.process(ctx); err != nil {
				action := c.Name
				if strings.HasSuffix(action, ".go") {
					action = "Hook"
				}
				log.Fatalln(color.RedString("%s: %s", action, err.Error()))
			}
		} else {
			log.Println(color.YellowString("invalid command: %s", cmd))
		}
	}
}

func (c *Command) process(ctx context.Context) (context.Context, error) {
	var err error
	for _, action := range Actions(c.Name) {
		if err = action.Do(ctx, c); err != nil {
			break
		}
	}
	return ctx, err //nolint:wrapcheck
}

func (c *Command) Stacks() []string {
	return c.stack
}
