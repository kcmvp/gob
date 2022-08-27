package cmd

import (
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type RunTestSuite struct {
	suite.Suite
	root, current string
}

func TestName(t *testing.T) {
	suite.Run(t, new(RunTestSuite))
}

func (s *RunTestSuite) SetupTest() {
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

//func (s *RunTestSuite) TestRunTest() {
//	b := bytes.NewBufferString("")
//	runCmd.SetOut(b)
//	rootCmd.SetArgs([]string{"run", "test"})
//	err := rootCmd.Execute()
//	require.Error(s.T(), err)
//}
