package builder

import (
	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type TestProject struct {
	root string
}

func (t *TestProject) GitHome() string {
	return filepath.Join(t.RootDir(), git.GitDirName)
}

func (t *TestProject) RootDir() string {
	return t.root
}

func (t *TestProject) ScriptDir() string {
	return filepath.Join(os.TempDir(), "scripts")
}

func (t *TestProject) TargetDir() string {
	return filepath.Join(os.TempDir(), "target")
}

func NewTestProject() *TestProject {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return &TestProject{
		root: filepath.Dir(dir),
	}
}

type ActionTestSuite struct {
	suite.Suite
	*Builder
	context.Context
}

func (s *ActionTestSuite) SetupSuite() {
	builder := NewBuilderWith(NewTestProject())
	s.Builder = builder
	s.Context = context.WithValue(context.Background(), CtxKeyBuilder, builder)
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

func (s *ActionTestSuite) TestBuilderAction() {
	err := os.RemoveAll(s.Builder.ScriptDir())
	require.NoError(s.T(), err)
	require.NoDirExists(s.T(), s.Builder.ScriptDir())
	if c, ok := commandMap()["builder"]; ok {
		_, err = c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.ScriptDir())
		require.FileExists(s.T(), filepath.Join(s.Builder.ScriptDir(), c.Output[0]))
		require.Equal(s.T(), len(c.Stacks()), 2)
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}
func (s *ActionTestSuite) TestGitHookAction() {
	err := os.RemoveAll(s.Builder.ScriptDir())
	require.NoError(s.T(), err)
	require.NoDirExists(s.T(), s.Builder.ScriptDir())
	if c, ok := commandMap()["gitHook"]; ok {
		_, err = c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.ScriptDir())
		for _, f := range c.Output {
			require.FileExists(s.T(), filepath.Join(s.Builder.ScriptDir(), f))
		}
		require.Equal(s.T(), len(c.Stacks()), 2)
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}

/*
func (s *ActionTestSuite) TestCleanActionNotExists() {
	err := os.RemoveAll(s.Builder.TargetDir())
	require.NoError(s.T(), err)
	if c, ok := commandMap()["clean"]; ok {
		_, err = c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.TargetDir())
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}

func (s *ActionTestSuite) TestCleanActionExists() {
	err := os.MkdirAll(s.Builder.TargetDir(), os.ModePerm)
	require.NoError(s.T(), err)
	if c, ok := commandMap()["clean"]; ok {
		_, err = c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.TargetDir())
	} else {
		require.Fail(s.T(), "invalid command")
	}
}
*/

func (s *ActionTestSuite) TestLintAction() {
	if c, ok := commandMap()["lint"]; ok {
		_, err := c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.TargetDir())
		for _, f := range c.Output {
			_, err = os.Stat(filepath.Join(s.Builder.TargetDir(), f))
			require.NoError(s.T(), err)
		}
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}

func (s *ActionTestSuite) TestTestAction() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	if c, ok := commandMap()["test"]; ok {
		_, err := c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.TargetDir())
		//for _, f := range c.Output {
		//	_, err = os.Stat(filepath.Join(s.Builder.TargetDir(), f))
		//	require.NoError(s.T(), err)
		//}
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}
