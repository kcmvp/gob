package scaffolds

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
		boot.PreCommit:   {createDirAction, cleanAction, lintAction},
		boot.CommitMsg:   {createDirAction, commitMsgAction, testAction},
		boot.PrePush:     {createDirAction, cleanAction, testAction},
		boot.InitBuilder: {createDirAction, initBuilder},
		boot.InitHook:    {createDirAction, initHook},
		boot.InitLinter:  {createDirAction, initLinter},
		boot.InitList:    {listStacks},
		boot.Clean:       {cleanAction, initHook},
		boot.Lint:        {createDirAction, initHook, lintAction},
		boot.Test:        {createDirAction, initHook, testAction},
		boot.Build:       {createDirAction, initHook, testAction, buildAction},
		// @todo refactor #68, this command will show the history data in console
		boot.Report: {createDirAction, initHook, lintAction, testAction, reportAction},
	}
}
