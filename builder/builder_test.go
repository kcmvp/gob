package builder

import (
	"fmt"
	"github.com/kcmvp/gob/infra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

type BuilderTestSuite struct {
	suite.Suite
	builder *Builder
}

func (bs *BuilderTestSuite) SetupSuite() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	root := filepath.Dir(filepath.Dir(filename))
	bs.builder = NewBuilder(root)
}

func TestBuilderSuit(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

func (bs *BuilderTestSuite) TestSortWithPreCommitHook() {
	actions := sort(preCommitHook)
	require.Equal(bs.T(), actions, []Action{SetupGitHook, Lint, Test, preCommitHook})
}

func (bs *BuilderTestSuite) TestSortWithCommitMsgHook() {
	actions := sort(commitMsgHook)
	require.Equal(bs.T(), actions, []Action{SetupGitHook, commitMsgHook})
}

func (bs *BuilderTestSuite) TestSortWithPrePushHook() {
	actions := sort(prePushHook)
	require.Equal(bs.T(), actions, []Action{SetupGitHook, Test, prePushHook})
}

func (bs *BuilderTestSuite) TestPreCommitHook() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	bs.builder.Run(SetupGitHook, Clean, Lint, Test)
	// check lint report
	_, err := os.Stat(filepath.Join(bs.builder.targetDir, lintersOut))
	require.NoError(bs.T(), err)
	for s, g := range infra.Hooks() {
		gof := fmt.Sprintf("%s.go", g)
		path, err := filepath.Abs(filepath.Join(bs.builder.scriptDir, gof))
		require.NoError(bs.T(), err)
		_, err = os.Stat(path)
		require.NoError(bs.T(), err)

		path, err = filepath.Abs(filepath.Join(bs.builder.root, ".git", "hooks", s))
		require.NoError(bs.T(), err)
		info, err := os.Stat(path)
		require.NoError(bs.T(), err)
		require.True(bs.T(), time.Now().Nanosecond()/1e6-info.ModTime().Nanosecond()/1e6 < 1000)

	}
	// check test report
	_, err = os.Stat(filepath.Join(bs.builder.targetDir, "cover.out"))
	require.NoError(bs.T(), err)
	_, err = os.Stat(filepath.Join(bs.builder.targetDir, "package.out"))
	require.NoError(bs.T(), err)
	_, err = os.Stat(filepath.Join(bs.builder.targetDir, "coverage.json"))
	require.NoError(bs.T(), err)
}

func (bs *BuilderTestSuite) TestCreateDirs() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	for _, action := range []Action{Lint, Test, Build} {
		bs.builder.Run(Clean)
		_, err := os.Stat(bs.builder.targetDir)
		require.Error(bs.T(), err, "target dir should be deleted")

		bs.builder.Run(action)
		_, err = os.Stat(bs.builder.targetDir)
		require.NoError(bs.T(), err, "target dir should be created")
	}
}
