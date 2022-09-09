package boot

import "github.com/samber/lo"

type Command string

const (
	None         Command = "_"
	Clean        Command = "clean"
	Build        Command = "build"
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

func ToCommands(commands ...string) []Command {
	return lo.Map(commands, func(c string, _ int) Command {
		return Command(c)
	})
}

func (command Command) ValidFlags() []string {
	flagMap := map[Command][]string{
		SetupBuilder: []string{},
		SetupHook:    []string{},
		SetupLinter:  []string{"version"},
		Clean:        []string{"-cache", "-testcache", "-modcache", "-fuzzcache"},
		Lint:         []string{"all"},
		Test:         []string{},
		Build:        []string{},
	}
	return flagMap[command]
}
