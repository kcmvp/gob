package boot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/color"
	"github.com/samber/lo"
)

type (
	Action func(project Project, command Command) error
)

var executor *Executor

type Executor struct {
	flags map[string]interface{}
	ctx   context.Context
}

func init() {
	executor = &Executor{
		flags: map[string]interface{}{},
		ctx:   context.Background(),
	}
}

func AllFlags(command Command) []string {
	prefix := fmt.Sprintf("%s.", command.Name())
	return lo.FilterMap(lo.Keys(executor.flags), func(flag string, _ int) (string, bool) {
		if strings.HasPrefix(flag, prefix) {
			return strings.Split(flag, prefix)[1], true
		}
		return flag, false
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

func SaveExecCtx(command Command, value string) {
	executor.ctx = context.WithValue(executor.ctx, command.CtxKey(), value)
}

func GetExecCtx(command Command) string {
	return executor.ctx.Value(command.CtxKey()).(string)
}

func Run(project Project, commands ...Command) error {
	var ccs []Command
	if project.Initializer() != None {
		ccs = append(ccs, project.Initializer())
		h := strings.TrimRight(project.Initializer().Name(), ".go")
		log.Printf("Triggered by %s\n", strings.ReplaceAll(h, "_", "-"))
	} else {
		ccs = commands
	}
	lo.ForEach(ccs, func(c Command, _ int) {
		if !lo.Contains(lo.Keys(project.Mapper()), c) {
			log.Fatalln(color.RedString("Invalid command: %s for %T", c, project))
		}
	})

	var err error
	lo.EveryBy(ccs, func(command Command) bool {
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
