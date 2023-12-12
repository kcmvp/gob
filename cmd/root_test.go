package cmd

import (
	"bytes"
	"github.com/kcmvp/gob/internal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestBuilder_Build(t *testing.T) {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		return
	}
	os.Setenv("callFromTest", "1")
	os.Chdir(internal.CurProject().Root())
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"test"})
	err := rootCmd.Execute()
	require.NoError(t, err)
}
