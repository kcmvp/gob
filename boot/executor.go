package boot

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"log"
	"math"
	"strings"
)

type Executor struct {
	*viper.Viper
}

func NewExecutor() *Executor {
	executor := &Executor{
		viper.New(),
	}
	return executor
}

func (executor *Executor) Run(project Project, commands ...string) error {
	promote := false
	lo.ForEach(executor.AllKeys(), func(v string, _ int) {
		if !strings.Contains(v, ".") {
			log.Println(color.YellowString("Flag %s does not have command code as prefix", v))
			promote = true
		}
	})
	if promote {
		log.Println(color.YellowString("Flag should follow {command code}.{flag} format"))
	}
	var cc []string
	starter := project.Initializer()
	if len(starter) > 0 {
		cc = append(cc, starter)
	} else {
		cc = commands
	}
	for _, command := range cc {
		for _, action := range project.Mapper(command) {
			err := action(project, command, executor.Flags(command)...)
			if err != nil {
				log.Println(color.RedString("Failed to execute the command %s:%s", command, err.Error()))
				return err
			}
		}
	}

	return nil
}

func (executor *Executor) Flags(command string) []string {
	prefix := fmt.Sprintf("%s.", command)
	args := lo.FilterMap(executor.AllKeys(), func(k string, _ int) (string, bool) {
		if strings.HasPrefix(k, prefix) {
			switch v := executor.Get(k).(type) {
			case bool:
				if v {
					return "", true
				}
			case string, int:
				f1 := lo.Substring(k, len(prefix), math.MaxInt)
				return fmt.Sprintf("%s %s", f1, v), true
			}
			return "", false
		} else {
			return "", false
		}
	})
	if len(args) == 0 {
		log.Println(color.YellowString("No flags for command: %s", command))
	}
	return args
}
