package cmd

import (
	"bytes"
	"errors"
	"fmt"
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
	builder *builder.Builder
}

func (s *CmdTestSuite) SetupSuite() {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	for {
		if _, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
			os.Chdir(root)
			s.builder = builder.NewBuilder(root)
			break
		} else {
			root = filepath.Dir(root)
		}
	}
}

func (s *CmdTestSuite) TearDownSuite() {
	// revert all the changes in th file
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) TestSetupBuilder() {
	builder := filepath.Join(s.builder.ScriptDir(), "builder.go")
	os.Remove(builder)
	require.NoFileExists(s.T(), builder)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", "builder"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	rootCmd.SetArgs([]string{"setup", "builder"})
	err = rootCmd.Execute()
	require.NoError(s.T(), err)
	require.FileExists(s.T(), builder)
}

func (s *CmdTestSuite) TestSetupHook() {

	for k, v := range builder.HookMap() {
		gf := filepath.Join(s.builder.ScriptDir(), fmt.Sprintf("%s.go", k))
		err := os.Remove(gf)
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
		err = os.Remove(filepath.Join(s.builder.GitHome(), "hooks", v))
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
	}

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", "githook"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)

	for k, v := range builder.HookMap() {
		f := filepath.Join(s.builder.ScriptDir(), fmt.Sprintf("%s.go", k))
		require.FileExists(s.T(), f)
		f = filepath.Join(s.builder.GitHome(), "hooks", v)
		require.FileExists(s.T(), f)
	}

}

func (s *CmdTestSuite) TestSetupLint() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", "linter", "-v", "v1.49.0"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.builder.Config().GetString("toolset.golangci-lint"), "v1.49.0")
}

func (s *CmdTestSuite) TestRunLint() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"run", "lint"})
	err := rootCmd.Execute()
	require.Error(s.T(), err)

	// @todo flags
}

func (s *CmdTestSuite) TestDummy() {
	require.Equal(s.T(), 1, 2)
}

//@todo test case with flags
//@todo lint test with flags
