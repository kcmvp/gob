package cmd

import (
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActions(t *testing.T) {
	assert.Equal(t, 3, len(execValidArgs))
	assert.True(t, lo.Every(execValidArgs, []string{internal.CommitMsgCmd, internal.PreCommitCmd, internal.PreCommitCmd}))
}

func TestCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no match", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"first match", []string{internal.CommitMsgCmd, lo.RandomString(10, lo.LettersCharset)}, false},
		{"second match", []string{lo.RandomString(10, lo.LettersCharset), "msghook"}, true},
		{"more than 3", []string{internal.CommitMsgCmd, lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset),
			lo.RandomString(10, lo.AlphanumericCharset)},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := execCmd.Args(execCmd, test.args)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}
