package boot

import (
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	scriptDir = "scripts"
	targetDir = "target"
)

type Mapper func() map[Command][]Action

type Project interface {
	GitHome() string
	ScriptDir() string
	TargetDir() string
	RootDir() string
	Config() *viper.Viper
	SaveConfig(key, value string)
	Mapper() map[Command][]Action
	Initializer() Command
}

var _ Project = (*DefaultProject)(nil)

type DefaultProject struct {
	root        string
	cfg         *viper.Viper
	initializer Command
	mapper      Mapper
}

func (project *DefaultProject) Mapper() map[Command][]Action {
	return project.mapper()
}

func (project *DefaultProject) Initializer() Command {
	return project.initializer
}

func (project *DefaultProject) GitHome() string {
	dir := filepath.Join(project.RootDir(), git.GitDirName)
	return dir
}

func (project *DefaultProject) ScriptDir() string {
	return filepath.Join(project.RootDir(), scriptDir)
}

func (project *DefaultProject) TargetDir() string {
	return filepath.Join(project.RootDir(), targetDir)
}

func (project *DefaultProject) RootDir() string {
	return project.root
}

// NewProject @todo optimize project initialization
func NewProject(root string, mapper Mapper) DefaultProject {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		log.Fatalln(color.RedString("can not find go.mod in %s", root))
	} else {
		if _, err := modfile.Parse("go.mod", data, nil); err != nil {
			log.Fatalln(color.RedString("invalid mod file %v", err))
		}
	}
	project := DefaultProject{
		root,
		viper.New(),
		None,
		mapper,
	}
	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(2, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	more := true
	initializers := []Command{PreCommit, CommitMsg, PrePush}
	for more {
		frame, more = frames.Next()
		c, ok := lo.Find(initializers, func(command Command) bool {
			return filepath.Base(frame.File) == string(command)
		})
		if ok {
			project.initializer = c
			h := strings.TrimRight(string(c), ".go")
			log.Printf("Hook %s is triggered \n", strings.ReplaceAll(h, "_", "-"))
			break
		}
	}
	return project
}

func (project *DefaultProject) Config() *viper.Viper {
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

func (project *DefaultProject) SaveConfig(key, value string) {
	project.cfg.Set(key, value)
	err := project.cfg.WriteConfigAs(filepath.Join(project.RootDir(), "application.yml"))
	if err != nil {
		log.Println(color.RedString("Failed to save the configuration %s", err.Error()))
	}
}
