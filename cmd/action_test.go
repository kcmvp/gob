package cmd

import (
	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ActionTestSuite struct {
	suite.Suite
	binary string
}

func TestActionSuite(t *testing.T) {
	suite.Run(t, &ActionTestSuite{
		binary: filepath.Join(internal.CurProject().Target(), lo.If(internal.Windows(), "gob.exe").Else("gob")),
	})
}

func (suite *ActionTestSuite) SetupSuite() {
	os.Remove(suite.binary)
}

func (suite *ActionTestSuite) TestActionBuild() {
	err := buildAction(nil)
	assert.NoError(suite.T(), err)
	_, err = os.Stat(suite.binary)
	assert.NoError(suite.T(), err)
}

func (suite *ActionTestSuite) TestBeforeExecution() {
	actions := lo.Filter(buildActions(), func(item Action, _ int) bool {
		return !strings.Contains(item.A, "_")
	})
	args := lo.Map(actions, func(item Action, _ int) string {
		return item.A
	})
	for _, arg := range args {
		assert.NoError(suite.T(), beforeExecution(nil, arg))
	}
}

func (suite *ActionTestSuite) TestBuiltInActions() {
	assert.Equal(suite.T(), 4, len(buildActions()))
	assert.Equal(suite.T(), []string{"build", "clean", "test", "after_test"}, lo.Map(buildActions(), func(item Action, index int) string {
		return item.A
	}))
}

func (suite *ActionTestSuite) TestExecute() {
	_ = os.Chdir(internal.CurProject().Root())
	err := execute(builderCmd, "build")
	assert.NoError(suite.T(), err)
	err = execute(builderCmd, "build1")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), color.RedString("can not find command %s", "build1"))
}

func (suite *ActionTestSuite) TestCoverage() {
	_, err1 := os.Stat(filepath.Join(internal.CurProject().Target(), "cover.out"))
	err2 := coverReport(nil, "")
	assert.True(suite.T(), (err1 == nil) == (err2 == nil))
}

func (suite *ActionTestSuite) TestSetupActions() {
	assert.Equal(suite.T(), 1, len(setupActions()))
}

func (suite *ActionTestSuite) TestSetupVersion() {
	err := setupVersion(nil, "")
	assert.NoError(suite.T(), err)
	version := filepath.Join(internal.CurProject().Root(), "infra", "version.go")
	os.Remove(version)
	_, err = os.Stat(version)
	assert.Error(suite.T(), err)
	err = setupVersion(nil, "")
	assert.NoError(suite.T(), err)
	_, err = os.Stat(version)
	assert.NoError(suite.T(), err)
}
