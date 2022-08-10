package builder

import (
	"errors"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

type CommitMessageSuit struct {
	suite.Suite
	project *Project
}

func (suit *CommitMessageSuit) SetupTest() {
	suit.project = NewProject(DefaultHookCfg())
	os.Chdir(suit.project.ModuleDir())
}

func TestMessageSuit(t *testing.T) {
	suite.Run(t, new(CommitMessageSuit))
}

func (suit *CommitMessageSuit) TestCommitMessageFormat() {
	errR := errors.New(color.RedString("commit message must follow %s", suit.project.gitHook.cfg.MsgPattern))
	tests := []struct {
		name   string
		args   []string
		result error
	}{
		{
			name:   "happyflow",
			args:   []string{"#123:1234567890"},
			result: nil,
		},
		{
			name:   "<10",
			args:   []string{"#123:123456789"},
			result: errR,
		},
		{
			name:   "no#",
			args:   []string{"123:1234567890"},
			result: errR,
		},
		{
			name:   "no:",
			args:   []string{"#1231234567890"},
			result: errR,
		},
		{
			name:   "length of # is bigger than 7",
			args:   []string{"#12345678:1234567890"},
			result: errR,
		},
		{
			name:   "space before :",
			args:   []string{"#123456 :1234567890"},
			result: errR,
		},
		{
			name:   "one space after :",
			args:   []string{"#123456: 1234567890"},
			result: nil,
		},
		{
			name:   "two spaces after :",
			args:   []string{"#123456:  1234567890"},
			result: errR,
		},
		{
			name:   "space between message#1",
			args:   []string{"#123456:123456789 0"},
			result: nil,
		},
		{
			name:   "space between message#2",
			args:   []string{"#123456: 1 2 3 4 5 6 7 890"},
			result: nil,
		},
		{
			name:   "space between message#3",
			args:   []string{"#123456: 123 4567 890"},
			result: nil,
		},
		{
			name:   "unique#1",
			args:   []string{"#123456: 你好你好你好你好你好"},
			result: nil,
		},
		{
			name:   "unique#2",
			args:   []string{"#123456: 你好 hello world!"},
			result: nil,
		},
		{
			name:   "newLine#2",
			args:   []string{"#123456: 你好 \nhello world!"},
			result: nil,
		},
	}
	for _, tt := range tests {
		suit.T().Run(tt.name, func(t *testing.T) {
			err := suit.project.gitHook.commitMessageBeforeScan(tt.args...)
			require.Equal(t, tt.result, err)
		})
	}
}
