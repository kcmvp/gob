package command

import (
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDependency(t *testing.T) {
	os.Chdir(artifact.CurProject().Root())
	_, err := dependencyTree()
	assert.NoError(t, err)
	//tree.VisitAll(func(item *treeprint.Node) {
	//	contains := lo.ContainsBy(artifact.CurProject().Dependencies(), func(dep *modfile.Require) bool {
	//		return strings.Contains(fmt.Sprintf("%s", item.Value), dep.Mod.Path)
	//	})
	//	assert.True(t, contains)
	//})
}
