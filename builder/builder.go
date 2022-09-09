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
		boot.PreCommit:    {CleanAction, LintAction},
		boot.CommitMsg:    {CommitMsgAction, TestAction},
		boot.PrePush:      {CleanAction, TestAction},
		boot.SetupBuilder: {CreateDirAction, GenBuilder},
		boot.SetupHook:    {CreateDirAction, GenHook},
		boot.SetupLinter:  {CreateDirAction, SetupLinter},
		boot.Clean:        {CleanAction, GenHook},
		boot.Lint:         {CreateDirAction, GenHook, LintAction},
		boot.Test:         {CreateDirAction, GenHook, TestAction},
		boot.Build:        {CreateDirAction, GenHook, BuildAction},
	}
}
