package command

import (
	"fmt"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
	"golang.org/x/mod/modfile"
	"os"
	"strings"
	"testing"
)

func TestDependency(t *testing.T) {
	os.Chdir(artifact.CurProject().Root())
	tree, err := dependencyTree()
	assert.NoError(t, err)
	tree.VisitAll(func(item *treeprint.Node) {
		contains := lo.ContainsBy(artifact.CurProject().Dependencies(), func(dep *modfile.Require) bool {
			return strings.Contains(fmt.Sprintf("%s", item.Value), dep.Mod.Path)
		})
		assert.True(t, contains)
	})
	depCmd.RunE(nil, []string{""})
}
