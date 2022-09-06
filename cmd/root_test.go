package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CmdTestSuite struct {
	suite.Suite
	project *boot.Project
}

func (s *CmdTestSuite) SetupSuite() {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	for {
		if _, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
			os.Chdir(root)
			s.project = builder.NewBuilder(root)
			break
		} else {
			root = filepath.Dir(root)
		}
	}
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) TestGenBuilder() {
	builder := filepath.Join(s.project.ScriptDir(), "builder.go")
	os.Remove(builder)
	require.NoFileExists(s.T(), builder)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"gen", "builder"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	rootCmd.SetArgs([]string{"gen", "builder"})
	err = rootCmd.Execute()
	require.NoError(s.T(), err)
	require.FileExists(s.T(), builder)
}

func (s *CmdTestSuite) TestGenHook() {

	for k, v := range boot.HookMap() {
		gf := filepath.Join(s.project.ScriptDir(), fmt.Sprintf("%s.go", k))
		err := os.Remove(gf)
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
		err = os.Remove(filepath.Join(s.project.GitHome(), "hooks", v))
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
	}

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"gen", "githook"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)

	for k, v := range boot.HookMap() {
		f := filepath.Join(s.project.ScriptDir(), fmt.Sprintf("%s.go", k))
		require.FileExists(s.T(), f)
		f = filepath.Join(s.project.GitHome(), "hooks", v)
		require.FileExists(s.T(), f)
	}

}

func (s *CmdTestSuite) TestGenLint() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"gen", "linter", "-v", "v1.49.0"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.project.Config().GetString("toolset.golangci-lint"), "v1.49.0")
}

func (s *CmdTestSuite) TestRunLint() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"run", "lint"})
	err := rootCmd.Execute()
	v, _ := s.project.CtxValue("lint.issues").(int)
	if v > 0 {
		require.Error(s.T(), err)
	}
}
