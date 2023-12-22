package cmd

import (
	"bufio"
	"github.com/kcmvp/gb/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type InitializationTestSuite struct {
	suite.Suite
	gopath string
}

func (suite *InitializationTestSuite) SetupSuite() {
	filepath.WalkDir(filepath.Join(suite.gopath, "bin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(d.Name(), "golangci-lint") {
			os.Remove(path)
		}
		return nil
	})
}

func (suite *InitializationTestSuite) TearDownSuite() {
	filepath.WalkDir(internal.CurProject().Target(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == internal.CurProject().Configuration() {
			//os.Remove(path)
		}
		return nil
	})
}

func TestInitializationTestSuit(t *testing.T) {
	suite.Run(t, &InitializationTestSuite{
		gopath: os.Getenv("GOPATH"),
	})
}

func (suite *InitializationTestSuite) TestInitializeHook() {
	initFunc(nil, nil)
	_, err := os.Stat(internal.CurProject().Configuration())
	assert.NoError(suite.T(), err)
	file, _ := os.Open(internal.CurProject().Configuration())
	scanner := bufio.NewScanner(file)
	var hasHook, hasMsg, hasOnCommit, hasOnPush, hasLint, hasAlias bool
	for scanner.Scan() {
		line := scanner.Text()
		hasHook = hasHook || strings.HasPrefix(line, internal.GitHookKey)
		hasMsg = hasMsg || strings.Contains(line, internal.CommitMsg)
		hasOnCommit = hasOnCommit || strings.Contains(line, internal.PreCommit)
		hasOnPush = hasOnPush || strings.Contains(line, internal.PrePush)
		hasLint = hasLint || strings.Contains(line, golangCiLinter)
		hasAlias = hasAlias || strings.Contains(line, "alias: lint")
	}
	assert.NoError(suite.T(), scanner.Err())
	assert.True(suite.T(), hasHook)
	assert.True(suite.T(), hasMsg)
	assert.True(suite.T(), hasOnPush)
	assert.True(suite.T(), hasOnCommit)
	assert.True(suite.T(), hasLint)
	assert.True(suite.T(), hasAlias)
	var installed bool
	filepath.WalkDir(filepath.Join(suite.gopath, "bin"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if installed = strings.HasPrefix(d.Name(), "golangci-lint-"); installed {
			return filepath.SkipDir
		}
		return nil
	})
	assert.True(suite.T(), installed)
}
