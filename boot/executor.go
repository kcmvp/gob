package boot

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"log"
	"strings"
)

type (
	Action func(project Project, command Command) error
)

var executor *Executor

type Executor struct {
	flags map[string]interface{}
}

func init() {
	executor = &Executor{
		map[string]interface{}{},
	}
}

func AllKeys() []string {
	//@todo don't expose this method
	return lo.Keys(executor.flags)
}

func AllFlags(command Command) []string {
	return lo.FilterMap(lo.Keys(executor.flags), func(flag string, _ int) (string, bool) {
		return flag, strings.HasPrefix(flag, fmt.Sprintf("%s.", command.Name()))
	})
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
		log.Fatalln(color.RedString("Invalid flag '%s' for command: %s", flag, command))
	}
	executor.flags[flagName(command, flag)] = value
}

func Run(project Project, commands ...Command) error {

	var ccs []Command
	if project.Initializer() != None {
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
				err = fmt.Errorf("[%s]:%w", command, err)
			}
			return err == nil
		})
	})
	if err == nil {
		log.Println("Run commands successfully")
	}
	return err
}
