package boot

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"log"
)

type Command string

const (
	none         Command = "_"
	Clean        Command = "clean"
	Build        Command = "build"
	Test         Command = "test"
	Lint         Command = "lint"
	SetupBuilder Command = "builder"
	SetupLinter  Command = "linter"
	SetupHook    Command = "githook"
	PreCommit    Command = "pre_commit.go"
	CommitMsg    Command = "commit_msg.go"
	PrePush      Command = "pre_push.go"
)

func (command Command) Name() string {
	return string(command)
}

func ToCommand(commands ...string) []Command {

	return nil
}

func (command Command) ValidFlags() []string {
	flagMap := map[Command][]string{
		SetupBuilder: []string{},
		SetupHook:    []string{},
		SetupLinter:  []string{"version"},
		Clean:        []string{"-cache", "-testcache", "-modcache", "-fuzzcache"},
		Lint:         []string{},
		Test:         []string{},
		Build:        []string{},
	}
	return flagMap[command]
}

type (
	Action func(project Project, command Command) error
)

var executor *ExecutorContext

type ExecutorContext struct {
	flags map[string]interface{}
}

func init() {
	executor = &ExecutorContext{
		map[string]interface{}{},
	}
}

func AllKeys() []string {
	//@todo don't expose this method
	return lo.Keys(executor.flags)
}

func flagName(command Command, flag string) string {
	return fmt.Sprintf("%s.%s", command, flag)
}

func GetFlag[T any](command Command, flag string) T {
	v, _ := executor.flags[flagName(command, flag)].(T)
	return v
}

func BindFlag(command Command, flag string, value interface{}) {
	if !lo.Contains(command.ValidFlags(), flag) {
		log.Fatalln(color.RedString("Invalid flag: %s for command: %s", flag, command))
	}
	executor.flags[flagName(command, flag)] = value
}

func run(project Project, commands ...Command) error {

	var ccs []Command
	if project.Initializer() != none {
		ccs = append(ccs, project.Initializer())
	} else {
		ccs = commands
	}
	lo.ForEach(commands, func(c Command, _ int) {
		if !lo.Contains(lo.Keys(project.Mapper()), c) {
			log.Fatalln(color.RedString("Invalid command: %s for %T", c, project))
		}
	})

	var err error
	lo.EveryBy(commands, func(command Command) bool {
		return lo.EveryBy(project.Mapper()[command], func(action Action) bool {
			err = action(project, command)
			if err != nil {
				err = fmt.Errorf("failed to execute the command %s:%w", command, err)
			}
			return err == nil
		})
	})
	if err == nil {
		log.Println("Commands run successfully")
	}
	return err
}
