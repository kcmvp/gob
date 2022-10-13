package boot

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommandOptions(t *testing.T) {
	require.Equal(t, Clean.ValidFlags(), []string{"-cache", "-testcache", "-modcache", "-fuzzcache", "delete"})
	require.Equal(t, Lint.ValidFlags(), []string{"all"})
	require.Equal(t, SetupLinter.ValidFlags(), []string{"version"})
	require.Equal(t, string(None), "")
}

func TestCommandKey(t *testing.T) {
	tests := []struct {
		Name   string
		cmd    Command
		errKey string
		ctxKey string
	}{
		{
			"test",
			Test,
			"Err.test",
			"ctx.test",
		},
		{
			"build",
			Build,
			"Err.build",
			"ctx.build",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.errKey, test.cmd.ErrKey())
			require.Equal(t, test.ctxKey, test.cmd.CtxKey())
		})
	}
}

func TestCommandHook(t *testing.T) {
	tests := []struct {
		Name string
		cmd  Command
		hook string
	}{
		{
			"test",
			Test,
			"",
		},
		{
			"prepush",
			PrePush,
			"pre-push",
		},
		{
			"pre-commit",
			PreCommit,
			"pre-commit",
		},
		{
			"commit-msg",
			CommitMsg,
			"commit-ms",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.hook, test.cmd.Hook())
		})
	}
}
