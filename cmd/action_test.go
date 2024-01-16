package cmd

import (
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

//func (suite *ActionTestSuite) TearDownSuite() {
//	os.Remove(suite.binary)
//}

func (suite *ActionTestSuite) TestActionBuild() {
	err := buildAction(nil)
	assert.NoError(suite.T(), err)
	_, err = os.Stat(suite.binary)
	assert.NoError(suite.T(), err)
}

func (suite *ActionTestSuite) TestBeforeExecution() {
	actions := lo.Filter(builtinActions(), func(item Action, _ int) bool {
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
	assert.Equal(suite.T(), 4, len(builtinActions()))
	assert.Equal(suite.T(), []string{"build", "clean", "test", "after_test"}, lo.Map(builtinActions(), func(item Action, index int) string {
		return item.A
	}))
}
