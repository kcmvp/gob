package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

type GitHookTestSuite struct {
	suite.Suite
	testDir string
}

func (suite *GitHookTestSuite) SetupSuite() {
	os.RemoveAll(suite.testDir)
}

func (suite *GitHookTestSuite) TearDownSuite() {
	os.RemoveAll(suite.testDir)
}

func TestGitHookSuite(t *testing.T) {
	_, file := TestCallee()
	suite.Run(t, &GitHookTestSuite{
		testDir: filepath.Join(CurProject().Target(), file),
	})
}

func (suite *GitHookTestSuite) TestSetupHook() {
	_, err := os.Stat(filepath.Join(suite.testDir, "gob.yaml"))
	assert.NotNil(suite.T(), err)
	CurProject().SetupHooks(true)
	info1, err := os.Stat(filepath.Join(suite.testDir, "gob.yaml"))
	assert.NoError(suite.T(), err)
	executions := CurProject().Executions()
	assert.Equal(suite.T(), 3, len(executions))
	rs := lo.Every([]string{"commit-msg-hook", "pre-commit-hook", "pre-push-hook"}, lo.Map(executions, func(item Execution, _ int) string {
		return item.CmdKey
	}))
	assert.True(suite.T(), rs)
	hook := CurProject().GitHook()
	assert.NotEmpty(suite.T(), hook.CommitMsg)
	assert.Equal(suite.T(), []string([]string{"lint", "test"}), hook.PreCommit)
	assert.Equal(suite.T(), []string([]string{"test"}), hook.PrePush)
	CurProject().SetupHooks(false)
	info2, err := os.Stat(filepath.Join(suite.testDir, "gob.yaml"))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), info1.ModTime(), info2.ModTime())
}
