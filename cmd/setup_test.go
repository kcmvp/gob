package cmd

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetupArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			"no arg",
			[]string{},
			true,
		},
		{
			"two args",
			[]string{"list", "githook"},
			true,
		},
		{
			"non-exist arg",
			[]string{"abc"},
			true,
		},
		{
			"list",
			[]string{"list"},
			false,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := setupCmd.ValidateArgs(testCase.args)
			require.True(t, testCase.wantErr == (err != nil))
		})
	}
}
