package project

import (
	"fmt"
	"github.com/kcmvp/gob/core/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type PluginTestSuit struct {
	suite.Suite
}

func (suite *PluginTestSuit) BeforeTest(_, testName string) {
	fmt.Println(RootDir())
	s, _ := os.Open(filepath.Join(RootDir(), "testdata", "build.yaml"))
	root := filepath.Join(RootDir(), "target", fmt.Sprintf("project_PluginTestSuit_%s", testName))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "build.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
}

func (suite *PluginTestSuit) TearDownSuite() {
	prefix := strings.TrimRight(utils.TestEnv().MustGet(), "TearDownSuite")
	TearDownSuite(prefix)
}

func TestPluginTestSuit(t *testing.T) {
	suite.Run(t, &PluginTestSuit{})
}

func (suite *PluginTestSuit) TestPlugins() {
	plugins := Plugins()
	assert.Len(suite.T(), plugins, 5)
	withV := lo.Filter(plugins, func(item Plugin, index int) bool {
		return len(item.Url) > 0
	})
	assert.Len(suite.T(), withV, 2)
}
