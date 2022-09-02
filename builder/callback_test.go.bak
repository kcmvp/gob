package builder

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gob/infra"
	"github.com/looplab/fsm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var _ Buildable = (*TestProject)(nil)

type TestProject struct {
	root string
}

func (t *TestProject) GitHome() (string, error) {
	return filepath.Join(t.RootDir(), git.GitDirName), nil
}

func (t *TestProject) RootDir() string {
	return t.root
}

func (t *TestProject) ScriptDir() string {
	return os.TempDir()
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

type CallbackTestSuite struct {
	suite.Suite
	*Builder
	context.Context
}

func (ts *CallbackTestSuite) SetupSuite() {
	builder := NewBuilderWith(NewTestProject())
	ts.Builder = builder
	ts.Context = context.WithValue(context.Background(), CtxKeyBuilder, builder)
}

func (ts *CallbackTestSuite) SetupTest() {
	// delete go file
	filepath.WalkDir(ts.ScriptDir(), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			for _, f := range []string{"builder.go", "commit_msg.go", "pre_commit.go", "pre_push.go"} {
				if err = os.Remove(filepath.Join(ts.ScriptDir(), f)); err != nil {
					return err
				}
			}
		}
		return err
	})

}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, new(CallbackTestSuite))
}

func (ts *CallbackTestSuite) TestDirs() {
	require.Equal(ts.T(), filepath.Join(os.TempDir(), "target"), ts.TargetDir())
	require.Equal(ts.T(), os.TempDir(), ts.ScriptDir())
}

func (ts *CallbackTestSuite) TestGenBuilder() {
	_, err := os.Stat(filepath.Join(ts.ScriptDir(), "builder.go"))
	require.Error(ts.T(), err)
	genBuilderCallback(ts.Context, nil)
	_, err = os.Stat(filepath.Join(ts.ScriptDir(), "builder.go"))
	require.NoError(ts.T(), err)
}

func (ts *CallbackTestSuite) TestGenGitHookScripts() {
	ctx := context.WithValue(ts.Context, GenHook, true)
	genGitHookCallback(ctx, nil)
	gitHome, _ := ts.GitHome()
	for k, _ := range infra.Hooks() {
		path, err := filepath.Abs(filepath.Join(gitHome, "hooks", k))
		require.NoError(ts.T(), err)
		info, err := os.Stat(path)
		require.NoError(ts.T(), err)
		require.True(ts.T(), time.Now().Nanosecond()/1e6-info.ModTime().Nanosecond()/1e6 < 1000)
	}
}

func (ts *CallbackTestSuite) TestLint() {
	lintCallback(ts.Context, nil)
	for _, r := range actionOutputMap(Lint) {
		_, err := os.Stat(filepath.Join(ts.TargetDir(), r))
		require.NoError(ts.T(), err)
	}
}

func (ts *CallbackTestSuite) TestCallback() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	testCallback(ts.Context, nil)
	for _, r := range actionOutputMap(Test) {
		_, err := os.Stat(filepath.Join(ts.TargetDir(), r))
		require.NoError(ts.T(), err)
	}
}

func (ts *CallbackTestSuite) TestCreateDir() {
	os.Remove(ts.TargetDir())
	os.Remove(ts.ScriptDir())
	createDirCallback(ts.Context, &fsm.Event{Dst: string(Test)})
	_, err := os.Stat(ts.TargetDir())
	require.NoError(ts.T(), err)
	createDirCallback(ts.Context, &fsm.Event{Dst: string(GenGitHook)})
	_, err = os.Stat(ts.ScriptDir())
	require.NoError(ts.T(), err)
}
