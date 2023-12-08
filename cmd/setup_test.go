package cmd

import (
	"bytes"
	"github.com/kcmvp/gob/internal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {
	os.Chdir(internal.CurProject().Root())
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	//rootCmd.SetArgs([]string{"action", "--cache"})
	rootCmd.SetArgs([]string{"setup", "-l"})
	err := rootCmd.Execute()
	require.NoError(t, err)
}

//func TestSetup_GitHook(t *testing.T) {
//	os.Chdir(internal.CurProject().Root())
//	b := bytes.NewBufferString("")
//	rootCmd.SetOut(b)
//	//rootCmd.SetArgs([]string{"action", "--cache"})
//	rootCmd.SetArgs([]string{"setup", "githook"})
//	err := rootCmd.Execute()
//	require.NoError(t, err)
//}
