package boot

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"

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

type Action func(project *Project, action string) error

type Actions func(action string) []Action

type Project struct {
	*buildOption
	root      string
	scriptDir string
	targetDir string
	hook      string
	actions   Actions
	*viper.Viper
	ctx context.Context
}

func (project *Project) Run(cmds ...string) {
	project.RunCtx(context.Background(), cmds...)
}

func (project *Project) RunCtx(ctx context.Context, cmds ...string) {
	project.ctx = ctx
	for _, cmd := range cmds {
		for _, action := range project.actions(cmd) {
			err := action(project, cmd)
			if err != nil {
				log.Fatalln(color.RedString("faild to execute the command %s:%s", cmd, err.Error()))
			}
		}
	}
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

func NewProject(root string, actions Actions) *Project {
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
		actions,
		viper.New(),
		context.Background(),
	}

	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(2, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	more := true
	for more {
		frame, more = frames.Next()
		hook := filepath.Base(frame.File)
		if lo.Contains(hooks(), hook) {
			project.hook = hook
			break
		}
	}

	return project
}

func (project *Project) IsCommitHook() bool {
	return project.hook == "pre_commit.go" || project.hook == "pre_push.go"
}

func (project *Project) ToolVersion(cmd string) string {
	return ""
}

func hooks() []string {
	return []string{
		"pre_commit.go",
		"commit_msg.go",
		"pre_push.go",
	}
}
