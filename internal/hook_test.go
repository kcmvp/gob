package internal

import (
	"bufio"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type GitHookTestSuite struct {
	suite.Suite
	gopath string
	start  int64
}

func (suite *GitHookTestSuite) TearDownSuite() {
	filepath.WalkDir(CurProject().Target(), func(path string, d fs.DirEntry, err error) error {
		if strings.HasPrefix(d.Name(), "gb-") && strings.HasSuffix(d.Name(), ".yaml") {
			os.Remove(path)
		}
		return err
	})
}

func TestGitHookSuite(t *testing.T) {
	suite.Run(t, &GitHookTestSuite{
		gopath: os.Getenv("GOPATH"),
		start:  time.Now().UnixNano(),
	})
}

func (suite *GitHookTestSuite) TestGitHook() {
	hook := CurProject().GitHook()
	assert.True(suite.T(), len(hook.CommitMsg) == 0)
	assert.True(suite.T(), len(hook.PreCommit) == 0)
	assert.True(suite.T(), len(hook.PrePush) == 0)
}

func (suite *GitHookTestSuite) TestSetupHook() {
	CurProject().Setup(true)
	hook := CurProject().GitHook()
	assert.Empty(suite.T(), hook.CommitMsg)
	assert.Empty(suite.T(), hook.PreCommit)
	assert.Empty(suite.T(), hook.PrePush)
	f, _ := os.Open(CurProject().Configuration())
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	data := strings.Join(lines, "\n")
	reader := strings.NewReader(data)
	err := CurProject().viper.ReadConfig(reader)
	hook = CurProject().GitHook()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hook.CommitMsg)
	assert.Equal(suite.T(), []string{"lint", "test"}, hook.PreCommit)
	assert.NotEmpty(suite.T(), hook.PreCommit)
	assert.NotEmpty(suite.T(), hook.PrePush)
	hooks := lo.MapToSlice(HookScripts, func(key string, _ string) string {
		return key
	})
	for _, h := range hooks {
		_, err = os.Stat(filepath.Join(CurProject().HookDir(), h))
		assert.NoError(suite.T(), err)
	}
	// test executions
	executions := project.Executions()
	assert.Equal(suite.T(), 3, len(executions))
	// drop last line the corresponding file should be deleted as well
	data = strings.Join(lo.DropRight(lines, 3), "\n")
	reader = strings.NewReader(data)
	CurProject().viper.ReadConfig(reader)
	CurProject().Setup(false)
	for _, h := range hooks {
		_, err = os.Stat(filepath.Join(CurProject().HookDir(), h))
		assert.Equal(suite.T(), err == nil, h != PrePush)
	}

}
