package builder

import (
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"strings"
	"sync"
)

var _ boot.Project = (*Builder)(nil)

type Builder struct {
	boot.DefaultProject
	*buildOption
}

func HookMap() map[string]string {
	return lo.KeyBy([]string{"pre-commit", "commit-msg", "pre-push"}, func(v string) string {
		return strings.ReplaceAll(v, "-", "_")
	})
}

var (
	instance *Builder
	once     sync.Once
)

// NewBuilder @todo refactor system can detect the root directory automatically
func NewBuilder(root string) *Builder {
	once.Do(func() {
		instance = &Builder{
			boot.NewProject(root, mapper),
			defaultOption(),
		}
	})
	return instance
}

var mapper = func() map[boot.Command][]boot.Action {
	return map[boot.Command][]boot.Action{
		boot.PreCommit:    {cleanAction, lintAction},
		boot.CommitMsg:    {commitMsgAction, testAction},
		boot.PrePush:      {cleanAction, testAction},
		boot.SetupBuilder: {createDirAction, genBuilder},
		boot.SetupHook:    {createDirAction, genHook},
		boot.SetupLinter:  {createDirAction, setupLinter},
		boot.Clean:        {cleanAction, genHook},
		boot.Lint:         {createDirAction, genHook, lintAction},
		boot.Test:         {createDirAction, genHook, testAction},
		boot.Build:        {createDirAction, genHook, buildAction},
	}
}
