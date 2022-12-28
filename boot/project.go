package boot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"golang.org/x/mod/modfile"
)

const (
	scriptDir = "scripts"
	targetDir = "target"

	BuildCfg = "build"
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
	Mod() *modfile.File
	Git() *git.Repository
}

var _ Project = (*DefaultProject)(nil)

type DefaultProject struct {
	root        string
	cfg         *viper.Viper
	initializer Command
	mapper      Mapper
	mod         *modfile.File
	repo        *git.Repository
}

func (project *DefaultProject) Git() *git.Repository {
	return project.repo
}

func (project *DefaultProject) Mod() *modfile.File {
	return project.mod
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

var hookInspector Inspector[string] = func(frame string) string {
	commands := []Command{PreCommit, CommitMsg, PrePush}
	for _, command := range commands {
		if filepath.Base(frame) == string(command) {
			return string(command)
		}
	}
	return string(None)
}

var rootInspector Inspector[string] = func(dir string) string {
	// @todo need to check windows root directory
	// @todo windows root directory
	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			log.Printf("Project root directory is %s\n", dir)
			return dir
		}
		dir = filepath.Dir(dir)
	}
	log.Fatalln("Can't find project root")
	return dir
}

func NewProject(root string, mapper Mapper) DefaultProject {
	project := DefaultProject{
		root: rootInspector(root),
		// root:        root,
		cfg:         viper.New(),
		mapper:      mapper,
		initializer: Command(Inspect(hookInspector)),
	}
	if data, err := os.ReadFile(filepath.Join(project.root, "go.mod")); err == nil {
		if mod, err := modfile.Parse("go.mod", data, nil); err != nil {
			log.Fatalln(color.RedString("invalid mod file %v", err))
		} else {
			project.mod = mod
		}
	}
	// init git
	if repo, err := git.PlainOpen(project.RootDir()); err == nil {
		project.repo = repo
	}
	return project
}

func (project *DefaultProject) Config() *viper.Viper {
	project.cfg.SetConfigName(BuildCfg)
	project.cfg.SetConfigType("yml")
	project.cfg.AddConfigPath(project.RootDir())
	if err := project.cfg.ReadInConfig(); err != nil {
		var t1 viper.ConfigFileNotFoundError
		if ok := errors.Is(err, t1); ok {
			log.Fatalln(color.RedString("Failed to read configuration %s", err.Error()))
		}
	}
	return project.cfg
}

func (project *DefaultProject) SaveConfig(key, value string) {
	project.cfg.Set(key, value)
	err := project.cfg.WriteConfigAs(filepath.Join(project.RootDir(), fmt.Sprintf("%s.yml", BuildCfg)))
	if err != nil {
		log.Println(color.RedString("Failed to save the configuration %s", err.Error()))
	}
}
