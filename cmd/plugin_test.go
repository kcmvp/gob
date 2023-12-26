package cmd

import (
	"bytes"
	"github.com/kcmvp/gob/cmd/action"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"
var v7 = "golang.org/x/tools/cmd/digraph@v0.16.1"
var fiximports = "golang.org/x/tools/cmd/fiximports@v0.16.1"

func TestInstallPlugin(t *testing.T) {
	internal.CurProject().LoadSettings()
	cfg := internal.CurProject().Configuration()
	defer func() {
		os.Remove(cfg)
	}()
	os.Chdir(internal.CurProject().Root())
	b := bytes.NewBufferString("")
	builderCmd.SetOut(b)
	builderCmd.SetArgs([]string{"plugin", "install", v6, "-a=callvis", "-c=run"})
	err := builderCmd.Execute()
	assert.NoError(t, err)
	plugin, ok := lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v6)
	assert.Equal(t, "digraph", plugin.A)
	assert.Equal(t, "callvis", plugin.B)
	assert.Equal(t, "run", plugin.C)
	assert.Equal(t, v6, plugin.D)
	// install same plugin again
	builderCmd.SetArgs([]string{"plugin", "install", v6, "-a=callvis", "-c=run"})
	err = builderCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v6)
	assert.Equal(t, "digraph", plugin.A)
	assert.Equal(t, "callvis", plugin.B)
	assert.Equal(t, "run", plugin.C)
	assert.Equal(t, v6, plugin.D)
	// install same plugin with different version
	builderCmd.SetArgs([]string{"plugin", "install", v7, "-a=callvis7", "-c=run7"})
	err = builderCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v7
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v7)
	assert.Equal(t, "digraph", plugin.A)
	assert.Equal(t, "callvis7", plugin.B)
	assert.Equal(t, "run7", plugin.C)
	assert.Equal(t, v7, plugin.D)

	// install another plugin
	builderCmd.SetArgs([]string{"plugin", "install", fiximports, "-a=lint", "-c=lint-run"})
	err = builderCmd.Execute()
	assert.NoError(t, err)
	builderCmd.SetArgs([]string{"plugin"})
	err = builderCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == fiximports
	})
	assert.Equal(t, "fiximports", plugin.A)
	assert.Equal(t, "lint", plugin.B)
	assert.Equal(t, "lint-run", plugin.C)
	assert.Equal(t, fiximports, plugin.D)
	assert.Equal(t, 2, len(internal.CurProject().Plugins()))
}

func TestInstallPluginWithVersion(t *testing.T) {
	ver, err := action.LatestVersion("github.com/hhatto/gocloc/cmd/gocloc", "")
	assert.NoError(t, err)
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"no version", "github.com/hhatto/gocloc/cmd/gocloc", false},
		{"latest version", "github.com/hhatto/gocloc/cmd/gocloc@latest", false},
		{"incorrect version", "github.com/hhatto/gocloc/cmd/gocloc@abc", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ver1, err1 := install(test.url)
			assert.True(t, test.wantErr == (err1 != nil))
			if err1 == nil {
				assert.Equal(t, ver, ver1)
			}
		})
	}
}
