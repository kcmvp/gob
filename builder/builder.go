package builder

import (
	"github.com/kcmvp/gob/boot"
)

var builderActions = func(cmdName string) []boot.Action {
	acm := map[string][]boot.Action{
		"pre_commit.go": {cleanAction, testAction, lintAction},
		"commit_msg.go": {commitMsgAction},
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

func NewBuilder(root string) *boot.Project {
	return boot.NewProject(root, builderActions)
}
