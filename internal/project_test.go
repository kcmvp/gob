package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

var v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"
var v7 = "golang.org/x/tools/cmd/digraph@v0.16.1"

func TestInstallPlugin(t *testing.T) {
	CurProject().LoadSettings()
	cfg := CurProject().Config()
	defer func() {
		os.Remove(cfg)
	}()
	err := CurProject().InstallPlugin(v6, "callvis")
	assert.NoError(t, err)
	gopath := os.Getenv("GOPATH")
	_, name := NormalizePlugin(v6)
	_, err = os.Stat(filepath.Join(gopath, "bin", name))
	assert.NoError(t, err)
	plugin, ok := lo.Find(CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.True(t, ok)
	assert.Equal(t, "digraph", plugin.A)
	assert.Equal(t, "callvis", plugin.B)
	assert.Empty(t, plugin.C)
	assert.Equal(t, v6, plugin.D)

	// install same plugin again
	err = CurProject().InstallPlugin(v6, "callvis")
	assert.Error(t, err)

	// update the plugin
	err = CurProject().InstallPlugin(v7, "callvisv7", "run")
	assert.NoError(t, err)
	plugin, ok = lo.Find(CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v7
	})
	assert.True(t, ok)
	assert.Equal(t, 1, len(CurProject().Plugins()))
	assert.Equal(t, plugin.B, "callvisv7")
	assert.Equal(t, plugin.C, "run")
}
