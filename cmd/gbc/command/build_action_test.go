package command

import (
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gbc/artifact"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ActionTestSuite struct {
	suite.Suite
}

func TestActionSuite(t *testing.T) {
	suite.Run(t, &ActionTestSuite{})
}

func (suite *ActionTestSuite) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))
}

func (suite *ActionTestSuite) TestActionBuild() {
	err := buildAction(nil)
	assert.NoError(suite.T(), err)
	binary := filepath.Join(artifact.CurProject().Target(), lo.If(artifact.Windows(), "gbc.exe").Else("gbc"))
	_, err = os.Stat(binary)
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
	_ = os.Chdir(artifact.CurProject().Root())
	err := execute(rootCmd, "build")
	assert.NoError(suite.T(), err)
	err = execute(rootCmd, "build1")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), color.RedString("can not find command %s", "build1"))
}

func (suite *ActionTestSuite) TestCoverage() {
	err := coverReport(nil, "")
	assert.Errorf(suite.T(), err, "no cover.out")
	s, _ := os.Open(filepath.Join(artifact.CurProject().Root(), "testdata", "cover.out"))
	t, _ := os.Create(filepath.Join(artifact.CurProject().Target(), "cover.out"))
	io.Copy(t, s)
	s.Close()
	t.Close()
	_, err = os.Stat(filepath.Join(artifact.CurProject().Target(), "cover.out"))
	err = coverReport(nil, "")
	assert.NoError(suite.T(), err, "should generate test cover report successfully")
	_, err = os.Stat(filepath.Join(artifact.CurProject().Target(), "cover.html"))
	assert.NoError(suite.T(), err)

}

func (suite *ActionTestSuite) TestBuildAndClean() {
	target := artifact.CurProject().Target()
	err := buildAction(nil, "")
	assert.NoError(suite.T(), err)
	entry, err := os.ReadDir(target)
	assert.Truef(suite.T(), len(entry) > 0, "target should not empty")
	err = cleanAction(nil, "")
	assert.NoError(suite.T(), err)
	entry, err = os.ReadDir(target)
	assert.Truef(suite.T(), len(entry) == 0, "target should empty")

}
