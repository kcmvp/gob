package cmd

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActions(t *testing.T) {
	assert.Equal(t, 1, len(actions))
	assert.True(t, lo.Every(actions, []string{"msghook"}))
}

func TestCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"no match", []string{lo.RandomString(10, lo.LettersCharset)}, true},
		{"first match", []string{"msghook", lo.RandomString(10, lo.LettersCharset)}, false},
		{"second match", []string{lo.RandomString(10, lo.LettersCharset), "msghook"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := execCmd.Args(execCmd, test.args)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}
