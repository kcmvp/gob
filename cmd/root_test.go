package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CmdTestSuite struct {
	suite.Suite
	builder *boot.Project
	l       sync.Mutex
	ctx     context.Context
}

func (s *CmdTestSuite) SetupSuite() {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	for {
		if _, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
			os.Chdir(root)
			s.builder = boot.NewProject(root)
			break
		} else {
			root = filepath.Dir(root)
		}
	}
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) BeforeTest(suiteName, testName string) {
	//s.l.Lock()
	session := boot.NewSession()
	s.ctx = context.WithValue(context.Background(), CurrentSession, session)
	if testName == "TestSetupLint" {
		log.Printf("session id: %s\n", session.ID())
	}
}

func (s *CmdTestSuite) AfterTest(suiteName, testName string) {
	//s.l.Unlock()
}

func (s *CmdTestSuite) TestSetupBuilder() {
	builder := filepath.Join(s.builder.ScriptDir(), "builder.go")
	_, err := os.Stat(builder)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{setupCmd.Name(), boot.SetupBuilder.Name()})
	err = rootCmd.ExecuteContext(s.ctx)
	if err == nil {
		require.NoError(s.T(), err, "should create builder.go successfully")
	} else {
		require.ErrorIs(s.T(), err, os.ErrExist, "should get file exists error")
	}
	session := rootCmd.Context().Value(CurrentSession).(*boot.Session)
	require.Empty(s.T(), session.AllFlags(boot.SetupBuilder))
}

func (s *CmdTestSuite) TestGenConfig() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{boot.Generate.Name(), "config"})
	rootCmd.ExecuteContext(s.ctx)
	session := rootCmd.Context().Value(CurrentSession).(*boot.Session)
	require.NotEmpty(s.T(), session.AllFlags(boot.Generate))
}

func (s *CmdTestSuite) TestSetupHook() {

	for _, v := range boot.HookMap() {
		err := os.Remove(filepath.Join(s.builder.GitHome(), "hooks", v))
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
	}

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{setupCmd.Name(), boot.SetupHook.Name()})
	err := rootCmd.ExecuteContext(s.ctx)
	require.NoError(s.T(), err)

	for k, v := range boot.HookMap() {
		f := filepath.Join(s.builder.ScriptDir(), fmt.Sprintf("%s.go", k))
		require.FileExists(s.T(), f)
		f = filepath.Join(s.builder.GitHome(), "hooks", v)
		require.FileExists(s.T(), f)
	}
	session := rootCmd.Context().Value(CurrentSession).(*boot.Session)
	require.Empty(s.T(), session.AllFlags(boot.SetupBuilder))

}
func (s *CmdTestSuite) TestSetupLint() {
	test := struct {
		name  string
		flags []string
		expV  string
	}{
		"withVersion",
		[]string{setupCmd.Name(), boot.SetupLinter.Name(), "-v", "v1.49.0"},
		"v1.49.0",
	}
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs(test.flags)
	err := rootCmd.ExecuteContext(s.ctx)
	require.NoError(s.T(), err)
	//@todo command does not switch context
	//session := s.ctx.Value(CurrentSession).(*boot.Session)
	//require.Equal(s.T(), test.expV, session.GetFlagString(boot.SetupLinter, "version"))
}
