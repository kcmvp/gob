package builder

import (
	"fmt"
	"github.com/kcmvp/gob/boot"
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
	builder *boot.Project
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

func (bs *BuilderTestSuite) TestPreCommitHook() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	bs.builder.Run("gitHook", "clean", "lint", "test")
	// check lint report
	_, err := os.Stat(filepath.Join(bs.builder.TargetDir(), "golangci-lint.html"))
	require.NoError(bs.T(), err)
	for g, s := range HookMap() {
		g = fmt.Sprintf("%s.go", g)
		path, err := filepath.Abs(filepath.Join(bs.builder.ScriptDir(), g))
		require.NoError(bs.T(), err)
		_, err = os.Stat(path)
		require.NoError(bs.T(), err)

		path, err = filepath.Abs(filepath.Join(bs.builder.RootDir(), ".git", "hooks", s))
		require.NoError(bs.T(), err)
		info, err := os.Stat(path)
		require.NoError(bs.T(), err)
		require.True(bs.T(), time.Now().Nanosecond()/1e6-info.ModTime().Nanosecond()/1e6 < 1000)

	}
	// check test report
	_, err = os.Stat(filepath.Join(bs.builder.TargetDir(), "cover.out"))
	require.NoError(bs.T(), err)
	_, err = os.Stat(filepath.Join(bs.builder.TargetDir(), "package.out"))
	require.NoError(bs.T(), err)
	_, err = os.Stat(filepath.Join(bs.builder.TargetDir(), "coverage.json"))
	require.NoError(bs.T(), err)
}

func (bs *BuilderTestSuite) TestBuild() {
	bs.builder.Run("build")
	require.FileExists(bs.T(), filepath.Join(bs.builder.TargetDir(), "main"))
}
