package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/modfile"
)

type CmdTestSuit struct {
	suite.Suite
	root, current string
}

func (s *CmdTestSuit) SetupTest() {
	_, file, _, ok := runtime.Caller(0)
	s.current = filepath.Dir(file)
	if ok {
		s.root = filepath.Dir(file)
		for {
			if _, err := os.ReadFile(filepath.Join(s.root, "go.mod")); err == nil {
				os.Chdir(s.root)
				break
			} else {
				s.root = filepath.Dir(s.root)
			}
		}
	}
	os.RemoveAll("scripts")
	os.RemoveAll(".golangci.yml")
}

func (s *CmdTestSuit) BeforeTest(suiteName, testName string) {
	if strings.EqualFold(testName, "TestRootCmdNotInRoot") {
		os.Chdir(s.current)
	}
}

func (s *CmdTestSuit) AfterTest(suiteName, testName string) {
	if strings.Contains(testName, "TestGithookCmd") {
		for k := range supportedHooks() {
			os.RemoveAll(filepath.Join(".git", "hooks", k))
		}
	}
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuit))
}

func (s *CmdTestSuit) TestRootCmdNotInRoot() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	require.Errorf(s.T(), err, "Error: please run the command in the module root directory")
}

func (s *CmdTestSuit) TestRootCmd() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{})
	rootCmd.Execute()
	_, err := io.ReadAll(b)
	require.NoError(s.T(), err)
	v, ok := rootCmd.Context().Value(_ctxModFileKey).(*modfile.File)
	require.True(s.T(), ok)
	require.NotNil(s.T(), v)
}

func (s *CmdTestSuit) TestNonExists() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"Hello"})
	rootCmd.Execute()
	msg, err := io.ReadAll(b)
	fmt.Println(msg)
	require.NoError(s.T(), err)
}

func (s *CmdTestSuit) TestBuilderCmd() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"builder"})
	rootCmd.Execute()
	_, err := io.ReadAll(b)
	require.NoError(s.T(), err)
	require.FileExists(s.T(), filepath.Join("scripts", "builder.go"))
	require.FileExists(s.T(), ".golangci.yml")
}

func (s *CmdTestSuit) TestGithookCmd() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"githook"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	for k, v := range supportedHooks() {
		require.FileExists(s.T(), filepath.Join(".git", "hooks", k))
		require.FileExists(s.T(), filepath.Join(scriptDir, v))
	}
}

func (s *CmdTestSuit) TestGithookCmdMultiple() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"githook"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	for k, v := range supportedHooks() {
		require.FileExists(s.T(), filepath.Join(".git", "hooks", k))
		require.FileExists(s.T(), filepath.Join(scriptDir, v))
	}
	err = rootCmd.Execute()
	require.NoError(s.T(), err)
	// data, _ := io.ReadAll(b)
	// require.Contains(s.T(), string(data), "commit-msg exists")
	// require.Contains(s.T(), string(data), "pre-push exists")
	for k, v := range supportedHooks() {
		require.FileExists(s.T(), filepath.Join(".git", "hooks", k))
		require.FileExists(s.T(), filepath.Join(scriptDir, v))
	}
}
