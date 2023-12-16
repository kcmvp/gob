package cmd

import (
	"bytes"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var v6 = "github.com/ofabry/go-callvis@v0.6.1"
var v7 = "github.com/ofabry/go-callvis@v0.7.0"
var golanglint = "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2"

func TestInstallPlugin(t *testing.T) {
	internal.CurProject().LoadSettings()
	cfg := internal.CurProject().Config()
	defer func() {
		os.Remove(cfg)
	}()
	os.Chdir(internal.CurProject().Root())
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"plugin", "install", v6, "-a=callvis", "-c=run"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	plugin, ok := lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v6)
	assert.Equal(t, "go-callvis", plugin.A)
	assert.Equal(t, "callvis", plugin.B)
	assert.Equal(t, "run", plugin.C)
	assert.Equal(t, v6, plugin.D)
	// install same plugin again
	rootCmd.SetArgs([]string{"plugin", "install", v6, "-a=callvis", "-c=run"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v6)
	assert.Equal(t, "go-callvis", plugin.A)
	assert.Equal(t, "callvis", plugin.B)
	assert.Equal(t, "run", plugin.C)
	assert.Equal(t, v6, plugin.D)
	// install same plugin with different version
	rootCmd.SetArgs([]string{"plugin", "install", v7, "-a=callvis7", "-c=run7"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v7
	})
	assert.Truef(t, ok, "%s should be installed successsfully", v7)
	assert.Equal(t, "go-callvis", plugin.A)
	assert.Equal(t, "callvis7", plugin.B)
	assert.Equal(t, "run7", plugin.C)
	assert.Equal(t, v7, plugin.D)

	// install another plugin
	rootCmd.SetArgs([]string{"plugin", "install", golanglint, "-a=lint", "-c=lint-run"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	rootCmd.SetArgs([]string{"plugin"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	plugin, ok = lo.Find(internal.CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == golanglint
	})
	assert.Equal(t, "golangci-lint", plugin.A)
	assert.Equal(t, "lint", plugin.B)
	assert.Equal(t, "lint-run", plugin.C)
	assert.Equal(t, golanglint, plugin.D)
	assert.Equal(t, 2, len(internal.CurProject().Plugins()))
}
