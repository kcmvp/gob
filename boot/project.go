package boot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"golang.org/x/mod/modfile"
)

const (
	scriptDir = "scripts"
	targetDir = "target"
)

type Action func(project *Project, cmd string) error

type Mapper func(cmd string) []Action

type Project struct {
	*buildOption
	root      string
	scriptDir string
	targetDir string
	hook      string
	mapper    Mapper
	cfg       *viper.Viper
	*viper.Viper
	ctx context.Context
}

func (project *Project) Run(cmds ...string) error {
	return project.RunCtx(context.Background(), cmds...)
}

func (project *Project) RunCtx(ctx context.Context, cmds ...string) error {
	project.ctx = ctx
	if len(project.hook) > 0 {
		cmds = []string{project.hook}
	}
	for _, cmd := range cmds {
		for _, action := range project.mapper(cmd) {
			err := action(project, cmd)
			if err != nil {
				log.Println(color.RedString("Failed to execute the command %s:%s", cmd, err.Error()))
				return err
			}
		}
	}
	return nil
}

func (project *Project) GitHome() string {
	dir := filepath.Join(project.RootDir(), git.GitDirName)
	return dir
}

func (project *Project) ScriptDir() string {
	return project.scriptDir
}

func (project *Project) TargetDir() string {
	return project.targetDir
}

func (project *Project) RootDir() string {
	return project.root
}

func NewProject(root string, mapper Mapper) *Project {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		log.Fatalln(color.RedString("can not find go.mod in %s", root))
	} else {
		if _, err := modfile.Parse("go.mod", data, nil); err != nil {
			log.Fatalln(color.RedString("invalid mod file %v", err))
		}
	}
	project := &Project{
		defaultOption(),
		root,
		filepath.Join(root, scriptDir),
		filepath.Join(root, targetDir),
		"",
		mapper,
		viper.New(),
		viper.New(),
		context.Background(),
	}

	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(2, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	more := true
	hooks := lo.MapKeys(HookMap(), func(_ string, k string) string {
		return fmt.Sprintf("%s.go", k)
	})
	for more {
		frame, more = frames.Next()
		hook := filepath.Base(frame.File)
		if lo.Contains(lo.Keys(hooks), hook) {
			project.hook = hook
			break
		}
	}
	return project
}

func (project *Project) Config() *viper.Viper {
	project.cfg.SetConfigName("application")
	project.cfg.SetConfigType("yml")
	project.cfg.AddConfigPath(project.RootDir())
	if err := project.cfg.ReadInConfig(); err != nil {
		// application.yml does not exist at very beginning
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalln(color.RedString("Failed to read configuration %s", err.Error()))
		}
	}
	return project.cfg
}

func (project *Project) SaveConfig(key, value string) {
	project.cfg.Set(key, value)
	err := project.cfg.WriteConfigAs(filepath.Join(project.RootDir(), "application.yml"))
	if err != nil {
		log.Println(color.RedString("Failed to save the configuration %s", err.Error()))
	}
}

func (project *Project) SaveCtx(k string, v interface{}) {
	project.ctx = context.WithValue(project.ctx, k, v)
}

func (project *Project) CtxValue(key string) interface{} {
	return project.ctx.Value(key)
}

func (project *Project) TriggeredByHook() bool {
	return true
}

func HookMap() map[string]string {
	return lo.KeyBy([]string{"pre-commit", "commit-msg", "pre-push"}, func(v string) string {
		return strings.ReplaceAll(v, "-", "_")
	})
}
