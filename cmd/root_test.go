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

type CmdTestSuite struct {
	suite.Suite
	root, current string
}

func (s *CmdTestSuite) SetupTest() {
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
}

func (s *CmdTestSuite) BeforeTest(suiteName, testName string) {
	if strings.EqualFold(testName, "TestRootCmdNotInRoot") {
		os.Chdir(s.current)
	}
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) TestRootCmdNotInRoot() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	require.Errorf(s.T(), err, "Error: please run the command in the module root directory")
}

func (s *CmdTestSuite) TestRootCmd() {
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

func (s *CmdTestSuite) TestNonExists() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"Hello"})
	rootCmd.Execute()
	msg, err := io.ReadAll(b)
	fmt.Println(msg)
	require.NoError(s.T(), err)
}

func (s *CmdTestSuite) TestBuilderCmd() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"gen", "builder"})
	rootCmd.Execute()
	_, err := io.ReadAll(b)
	require.NoError(s.T(), err)
	require.FileExists(s.T(), filepath.Join("scripts", "builder.go"))
	require.FileExists(s.T(), ".golangci.yml")
}
