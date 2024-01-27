package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type GitHookTestSuite struct {
	suite.Suite
}

func TearDownSuite(prefix string) {
	filepath.WalkDir(os.TempDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
	filepath.WalkDir(filepath.Join(CurProject().Root(), "target"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
}

func (suite *GitHookTestSuite) TearDownSuite() {
	TearDownSuite("internal_hook_test_")
}

func TestGitHookSuite(t *testing.T) {
	suite.Run(t, &GitHookTestSuite{})
}

func (suite *GitHookTestSuite) TestSetupHook() {
	CurProject().SetupHooks(true)
	info1, err := os.Stat(filepath.Join(CurProject().Target(), "gob.yaml"))
	assert.NoError(suite.T(), err)
	executions := CurProject().Executions()
	assert.Equal(suite.T(), 3, len(executions))
	rs := lo.Every([]string{"commit-msg", "pre-commit", "pre-push"}, lo.Map(executions, func(item Execution, _ int) string {
		return item.CmdKey
	}))
	assert.True(suite.T(), rs)
	hook := CurProject().GitHook()
	assert.NotEmpty(suite.T(), hook.CommitMsg)
	assert.Equal(suite.T(), []string([]string{"lint", "test"}), hook.PreCommit)
	assert.Equal(suite.T(), []string([]string{"test"}), hook.PrePush)
	CurProject().SetupHooks(false)
	info2, err := os.Stat(filepath.Join(CurProject().Target(), "gob.yaml"))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info1.ModTime(), info2.ModTime())
}
