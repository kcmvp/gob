package cmd

import (
	"github.com/kcmvp/gob/internal"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestValidSetupArgs(t *testing.T) {
	assert.Equal(t, setupCmd.ValidArgs, []string{"version"})
}

func TestSetupVersion(t *testing.T) {
	version := filepath.Join(internal.CurProject().Root(), "infra", "version.go")
	os.Remove(version)
	_, err := os.Stat(version)
	assert.Error(t, err)
	builderCmd.SetArgs([]string{"setup", "version"})
	err = builderCmd.Execute()
	assert.NoError(t, err)
	_, err = os.Stat(version)
	assert.NoError(t, err)
}
