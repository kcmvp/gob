package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	require.NoError(s.T(), err)
}

func (s *CmdTestSuite) TestNonExists() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"Hello"})
	err := rootCmd.Execute()
	require.Error(s.T(), err)
}

//func (s *CmdTestSuite) TestBuilderCmd() {
//	b := bytes.NewBufferString("")
//	rootCmd.SetOut(b)
//	rootCmd.SetArgs([]string{"gen", "builder"})
//	rootCmd.Execute()
//	_, err := io.ReadAll(b)
//	require.NoError(s.T(), err)
//	require.FileExists(s.T(), filepath.Join("scripts", "builder.go"))
//	require.FileExists(s.T(), ".golangci.yml")
//}
