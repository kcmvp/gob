package boot

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/samber/lo"
)

type (
	Action func(session *Session, project Project, command Command) error
)

type Session struct {
	store map[string]interface{}
	id    string
}

func NewSession() *Session {
	return &Session{
		store: map[string]interface{}{},
		id:    strconv.FormatInt(time.Now().UnixNano(), 10),
	}
}

func (session *Session) Specified(name string) string {
	return fmt.Sprintf("%s.%s", name, session.id)
}

func (session *Session) ID() string {
	return session.id
}

func (session *Session) AllFlags(command Command) []string {
	prefix := fmt.Sprintf("%s.", command.Name())
	return lo.FilterMap(lo.Keys(session.store), func(flag string, _ int) (string, bool) {
		if strings.HasPrefix(flag, prefix) {
			return strings.Split(flag, prefix)[1], true
		}
		return flag, false
	})
}

func flagName(command Command, flag string) string {
	return fmt.Sprintf("%s.%s", command, flag)
}

func (session *Session) GetFlagBool(command Command, flag string) bool {
	v, _ := session.store[flagName(command, flag)].(bool)
	return v
}

func (session *Session) GetFlagString(command Command, flag string) string {
	v, _ := session.store[flagName(command, flag)].(string)
	return v
}

func (session *Session) BindFlag(command Command, flag string, value interface{}) {
	if !lo.Contains(command.ValidFlags(), flag) {
		log.Fatalln(color.RedString("Invalid flag '%s' for command: %s", flag, command))
	}
	session.store[flagName(command, flag)] = value
}

func (session *Session) SaveCtxValue(command Command, value string) {
	session.store[command.CtxKey()] = value
}

func (session *Session) CtxValue(command Command) string {
	return session.store[command.CtxKey()].(string)
}

func (session *Session) Run(project Project, commands ...Command) error {
	var ccs []Command
	if project.Initializer() != None {
		ccs = append(ccs, project.Initializer())
		log.Printf("Triggered by hook %s\n", project.Initializer().Hook())
	} else {
		ccs = commands
	}
	lo.ForEach(ccs, func(c Command, _ int) {
		if !lo.Contains(lo.Keys(project.Mapper()), c) {
			log.Fatalln(color.RedString("Invalid command: %s for %T", c, project))
		}
	})
	defer session.cleanup(project)
	var err error
	lo.EveryBy(ccs, func(command Command) bool {
		return lo.EveryBy(project.Mapper()[command], func(action Action) bool {
			err = action(session, project, command)
			if err != nil {
				log.Println(color.RedString("%s", err.Error()))
			}
			return err == nil
		})
	})
	if err == nil {
		log.Println("Run commands successfully")
	}
	return err
}

func (session *Session) cleanup(project Project) {
	err := filepath.WalkDir(project.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), session.ID()) {
			np := strings.TrimSuffix(path, fmt.Sprintf(".%s", session.ID()))
			err = os.Rename(path, np)
			if err != nil {
				return fmt.Errorf("failed to rename file %s to %s: %w", path, np, err)
			}
		}
		return err //nolint:wrapcheck
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println(color.RedString("Failed to clean up the project:%s", err.Error()))
	}
}
