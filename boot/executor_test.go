package boot

import (
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestExecutor_Flags(t *testing.T) {
	BindFlag("lint.all", true)
	BindFlag("lint.hell", "111")
	BindFlag("lint.changed", false)
	sorted := AllKeys()
	sort.Strings(sorted)
	require.Equal(t, []string{"lint.all", "lint.changed", "lint.hell"}, sorted)
	v1 := GetFlag[bool]("lint.all")
	require.Equal(t, v1, true)

	v2 := GetFlag[string]("lint.all")
	require.Equal(t, v2, "")
}
