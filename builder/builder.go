package builder

import (
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"strings"
)

var _ boot.Project = (*Builder)(nil)

type Builder struct {
	boot.DefaultProject
	*buildOption
}

var builderActions = func(cmdName string) []boot.Action {
	acm := map[string][]boot.Action{
		"pre_commit.go": {cleanAction, lintAction},
		"commit_msg.go": {commitMsgAction, testAction},
		"pre_push.go":   {cleanAction, testAction},
		"builder":       {createDirAction, genBuilder},
		"githook":       {createDirAction, getHook},
		"linter":        {createDirAction, setupLinter},
		"clean":         {cleanAction, getHook},
		"lint":          {createDirAction, getHook, lintAction},
		"test":          {createDirAction, getHook, testAction},
		"build":         {createDirAction, getHook, buildAction},
	}
	return acm[cmdName]
}

func HookMap() map[string]string {
	return lo.KeyBy([]string{"pre-commit", "commit-msg", "pre-push"}, func(v string) string {
		return strings.ReplaceAll(v, "-", "_")
	})
}

func NewBuilder(root string) *Builder {
	builder := &Builder{
		boot.NewProject(root, builderActions),
		defaultOption(),
	}
	return builder
}
