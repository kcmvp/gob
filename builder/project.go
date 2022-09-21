package builder

import (
	"strings"
	"sync"

	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
)

var _ boot.Project = (*Project)(nil)

var (
	instance *Project
	once     sync.Once
)

type Project struct {
	boot.DefaultProject
	*buildOption
}

func HookMap() map[string]string {
	return lo.KeyBy([]string{"pre-commit", "commit-msg", "pre-push"}, func(v string) string {
		return strings.ReplaceAll(v, "-", "_")
	})
}

func NewProject() *Project {
	once.Do(func() {
		instance = &Project{
			boot.NewProject(mapper),
			defaultOption(),
		}
	})
	return instance
}

var mapper = func() map[boot.Command][]boot.Action {
	return map[boot.Command][]boot.Action{
		boot.PreCommit:   {cleanAction, lintAction},
		boot.CommitMsg:   {commitMsgAction, testAction},
		boot.PrePush:     {cleanAction, testAction},
		boot.InitBuilder: {createDirAction, initBuilder},
		boot.InitHook:    {createDirAction, initHook},
		boot.InitLinter:  {createDirAction, initLinter},
		boot.Clean:       {cleanAction, initHook},
		boot.Lint:        {createDirAction, initHook, lintAction},
		boot.Test:        {createDirAction, initHook, testAction},
		boot.Build:       {createDirAction, initHook, buildAction},
		boot.Report:      {createDirAction, initHook, lintAction, testAction, reportAction},
	}
}
