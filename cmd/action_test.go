package cmd

import (
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
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

func (suite *ActionTestSuite) TestBuildClean() {
	err := buildAction(nil)
	assert.NoError(suite.T(), err)
	_, err = os.Stat(suite.binary)
	assert.NoError(suite.T(), err)
}
