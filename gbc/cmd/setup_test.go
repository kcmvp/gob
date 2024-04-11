package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidSetupArgs(t *testing.T) {
	assert.Equal(t, setupCmd.ValidArgs, []string{"version"})
}

//
//func TestSetupVersion(t *testing.T) {
//	version := filepath.Join(artifact.CurProject().Root(), "infra", "version.go")
//	os.Remove(version)
//	_, err := os.Stat(version)
//	assert.Error(t, err)
//	rootCmd.SetArgs([]string{"setup", "version"})
//	err = rootCmd.Execute()
//	assert.NoError(t, err)
//	_, err = os.Stat(version)
//	assert.NoError(t, err)
//}
