package cmd

import (
	"fmt"
	"github.com/kcmvp/gob/gbc/artifact"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/xlab/treeprint"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMod(t *testing.T) {
	os.Chdir(artifact.CurProject().Root())
	mod, _ := os.Open(filepath.Join(artifact.CurProject().Root(), "go.mod"))
	m, _, deps, err := parseMod(mod)
	assert.NoError(t, err)
	assert.Equal(t, m, "github.com/kcmvp/gob")
	assert.Equal(t, 15, len(lo.Filter(deps, func(item *lo.Tuple4[string, string, string, int], _ int) bool {
		return item.D == 1
	})))
	assert.Equal(t, 48, len(deps))
}

func TestDependency(t *testing.T) {
	os.Chdir(artifact.CurProject().Root())
	mod, _ := os.Open(filepath.Join(artifact.CurProject().Root(), "go.mod"))
	_, _, deps, _ := parseMod(mod)
	tree, err := dependencyTree()
	assert.NoError(t, err)
	tree.VisitAll(func(item *treeprint.Node) {
		contains := lo.ContainsBy(deps, func(dep *lo.Tuple4[string, string, string, int]) bool {
			return strings.Contains(fmt.Sprintf("%s", item.Value), fmt.Sprintf("%s", dep.A))
		})
		assert.True(t, contains)
	})
	depCmd.RunE(nil, []string{""})
}
