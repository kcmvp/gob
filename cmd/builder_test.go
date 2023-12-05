package cmd

import (
	"bytes"
	"github.com/kcmvp/gob/internal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestBuilder_Build(t *testing.T) {
	os.Chdir(internal.CurProject().Root())
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	//rootCmd.SetArgs([]string{"build", "--cache"})
	rootCmd.SetArgs([]string{"build"})
	err := rootCmd.Execute()
	require.NoError(t, err)
}
