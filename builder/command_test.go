package builder

import (
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestGetCommandMap(t *testing.T) {
	m := commandMap()
	require.Equal(t, 9, len(m))
	var names []string
	for s, _ := range m {
		names = append(names, s)
	}
	sort.Strings(names)
	require.Equal(t, names, []string{"build", "builder", "clean", "commit-msg.go",
		"gitHook", "lint", "pre-commit.go", "pre-push.go", "test"})

}

func TestChildren(t *testing.T) {
	children := Children("run")
	require.Equal(t, 4, len(children))
	s1 := []string{"clean", "test", "lint", "build"}
	sort.Strings(s1)
	sort.Strings(children)
	require.Equal(t, s1, children)

	children = Children("gen")
	require.Equal(t, 2, len(children))

}

func TestGetAction(t *testing.T) {
	require.Equal(t, len(GetActions("pre-commit.go")), 3)
	require.Equal(t, len(GetActions("commit-msg.go")), 1)
	require.Equal(t, len(GetActions("pre-push.go")), 2)
	require.Equal(t, len(GetActions("builder")), 2)
	require.Equal(t, len(GetActions("gitHook")), 2)
	require.Equal(t, len(GetActions("clean")), 1)
	require.Equal(t, len(GetActions("lint")), 2)
	require.Equal(t, len(GetActions("test")), 2)
	require.Equal(t, len(GetActions("build")), 2)
}
