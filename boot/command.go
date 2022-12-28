package boot

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type (
	Command string
	Mapper  func() map[Command][]Action
)

const (
	None         Command = ""
	Clean        Command = "clean"
	Build        Command = "build"
	Report       Command = "report"
	Test         Command = "test"
	Lint         Command = "lint"
	SetupBuilder Command = "builder"
	SetupLinter  Command = "linter"
	SetupHook    Command = "githook"
	SetupGitFlow Command = "gitflow"
	PreCommit    Command = "pre_commit.go"
	CommitMsg    Command = "commit_msg.go"
	PrePush      Command = "pre_push.go"
	Generate     Command = "gen"
)

func (command Command) Name() string {
	return string(command)
}

func (command Command) Hook() string {
	var hookName string
	switch command { //nolint:exhaustive
	case PreCommit, CommitMsg, PrePush:
		hookName = strings.TrimRight(strings.ReplaceAll(command.Name(), "_", "-"), ".go")
	default:
		hookName = ""
	}
	return hookName
}

func (command Command) CtxKey() string {
	return fmt.Sprintf("ctx.%s", command.Name())
}

func (command Command) ErrKey() string {
	return fmt.Sprintf("Err.%s", command.Name())
}

func ToCommands(commands ...string) []Command {
	return lo.Map(commands, func(c string, _ int) Command {
		return Command(c)
	})
}

func (command Command) ValidFlags() []string {
	flagMap := map[Command][]string{
		SetupBuilder: {},
		SetupHook:    {},
		SetupLinter:  {"version"},
		Clean:        {"-cache", "-testcache", "-modcache", "-fuzzcache", "delete"},
		Lint:         {"all"},
		Test:         {},
		Build:        {},
		Generate:     {"stack"},
	}
	return flagMap[command]
}

var mapper = func() map[Command][]Action {
	return map[Command][]Action{
		PreCommit:    {createDirAction, cleanAction, lintAction},
		CommitMsg:    {createDirAction, commitMsgAction, testAction},
		PrePush:      {createDirAction, cleanAction, testAction},
		SetupBuilder: {createDirAction, setupBuilder, getGob},
		SetupHook:    {createDirAction, setupHook, getGob},
		SetupLinter:  {createDirAction, setupLinter},
		SetupGitFlow: {createDirAction, setupGitFlow},
		Clean:        {cleanAction, setupHook},
		Lint:         {createDirAction, setupHook, lintAction},
		Test:         {createDirAction, setupHook, testAction},
		Build:        {createDirAction, setupHook, testAction, buildAction},
		// @todo refactor #68, this command will show the history data in console
		Report:   {createDirAction, setupHook, lintAction, testAction, reportAction},
		Generate: {generate},
	}
}
