package boot

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommandOptions(t *testing.T) {
	require.Equal(t, Clean.ValidFlags(), []string{"-cache", "-testcache", "-modcache", "-fuzzcache", "delete"})
	require.Equal(t, Lint.ValidFlags(), []string{"all"})
	require.Equal(t, SetupLinter.ValidFlags(), []string{"version"})
}
