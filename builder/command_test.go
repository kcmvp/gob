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
	require.Equal(t, names, []string{"build", "builder", "clean", "commit_msg.go",
		"gitHook", "lint", "pre_commit.go", "pre_push.go", "test"})

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
	require.Equal(t, 3, len(Actions("pre_commit.go")))
	require.Equal(t, 1, len(Actions("commit_msg.go")))
	require.Equal(t, 2, len(Actions("pre_push.go")))
	require.Equal(t, 2, len(Actions("builder")))
	require.Equal(t, 2, len(Actions("gitHook")))
	require.Equal(t, 2, len(Actions("clean")))
	require.Equal(t, 3, len(Actions("lint")))
	require.Equal(t, 3, len(Actions("test")))
	require.Equal(t, 3, len(Actions("build")))
}
