package boot

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"

	"github.com/go-git/go-git/v5"

	"github.com/fatih/color"
	"github.com/spf13/viper"
	"golang.org/x/mod/modfile"
)

const (
	scriptDir = "scripts"
	targetDir = "target"
	BuildCfg  = "build"
)

const DefaultHookMsg = `((^|[\s])(#[0-9]{1,7}))+:\s?(\S+\s?){10,}`

// HookMsgPattern @todo should simplify.
type HookMsgPattern string

type BuildOption struct {
	MinCoverage float64
	MaxCoverage float64
	MsgPattern  HookMsgPattern
}

func DefaultOption() *BuildOption {
	return &BuildOption{
		MinCoverage: 0.35,
		MaxCoverage: 0.90,
		MsgPattern:  DefaultHookMsg,
	}
}

type Project struct {
	root        string
	cfg         *viper.Viper
	initializer Command
	mapper      Mapper
	mod         *modfile.File
	repo        *git.Repository
	option      *BuildOption
}

func (project *Project) Git() *git.Repository {
	return project.repo
}

func (project *Project) Mod() *modfile.File {
	return project.mod
}

func (project *Project) Mapper() map[Command][]Action {
	return project.mapper()
}

func (project *Project) Initializer() Command {
	return project.initializer
}

func (project *Project) GitHome() string {
	dir := filepath.Join(project.RootDir(), git.GitDirName)
	return dir
}

func (project *Project) ScriptDir() string {
	return filepath.Join(project.RootDir(), scriptDir)
}

func (project *Project) TargetDir() string {
	return filepath.Join(project.RootDir(), targetDir)
}

func (project *Project) RootDir() string {
	return project.root
}

func (project *Project) Option() *BuildOption {
	return project.option
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

func NewProject(root string) *Project {
	project := &Project{
		root: rootInspector(root),
		// root:        root,
		cfg:         viper.New(),
		mapper:      mapper,
		initializer: Command(Inspect(hookInspector)),
		option:      DefaultOption(),
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

func (project *Project) Config() *viper.Viper {
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

func (project *Project) SaveConfig(key, value string) {
	project.cfg.Set(key, value)
	err := project.cfg.WriteConfigAs(filepath.Join(project.RootDir(), fmt.Sprintf("%s.yml", BuildCfg)))
	if err != nil {
		log.Println(color.RedString("Failed to save the configuration %s", err.Error()))
	}
}

func HookMap() map[string]string {
	return lo.KeyBy([]string{"pre-commit", "commit-msg", "pre-push"}, func(v string) string {
		return strings.ReplaceAll(v, "-", "_")
	})
}
