package cmd

import (
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type GenTestSuite struct {
	suite.Suite
	root, current string
}

func (s *GenTestSuite) SetupTest() {
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

func TestGenTestSuite(t *testing.T) {
	suite.Run(t, new(GenTestSuite))
}

//func (s *GenTestSuite) TestGenHook() {
//	generateHook()
//}
