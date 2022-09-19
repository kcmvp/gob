package boot

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

type Command string

const (
	None         Command = "_"
	Clean        Command = "clean"
	Build        Command = "build"
	Report       Command = "report"
	Test         Command = "test"
	Lint         Command = "lint"
	SetupBuilder Command = "builder"
	SetupLinter  Command = "linter"
	SetupHook    Command = "githook"
	PreCommit    Command = "pre_commit.go"
	CommitMsg    Command = "commit_msg.go"
	PrePush      Command = "pre_push.go"
)

func (command Command) Name() string {
	return string(command)
}

func (command Command) Hook() string {
	var hook string
	switch command { //nolint:exhaustive
	case PreCommit, CommitMsg, PrePush:
		hook = strings.TrimRight(strings.ReplaceAll(command.Name(), "_", "-"), ".go")
	default:
		hook = ""
	}
	return hook
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
	}
	return flagMap[command]
}
