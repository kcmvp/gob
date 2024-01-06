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

const golangCiLinter = "github.com/golangci/golangci-lint/cmd/golangci-lint"
const testsum = "gotest.tools/gotestsum"

type InitializationTestSuite struct {
	suite.Suite
	testDir string
	goPath  string
}

func TestInitializationTestSuit(t *testing.T) {
	_, dir := internal.TestCallee()
	suite.Run(t, &InitializationTestSuite{
		testDir: filepath.Join(internal.CurProject().Target(), dir),
		goPath:  internal.GoPath(),
	})
}

func (suite *InitializationTestSuite) TearDownSuite() {
	os.RemoveAll(suite.goPath)
}

func (suite *InitializationTestSuite) TestBuiltInPlugins() {
	plugins := builtinPlugins()
	assert.Equal(suite.T(), 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin internal.Plugin) bool {
		return plugin.Module() == "github.com/golangci/golangci-lint"
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "golangci-lint", plugin.Name())
	assert.Equal(suite.T(), "lint", plugin.Alias)
}

func (suite *InitializationTestSuite) TestInitializerFunc() {
	initializerFunc(nil, nil)
	internal.CurProject().LoadSettings()
	plugins := internal.CurProject().Plugins()
	assert.Equal(suite.T(), 2, len(plugins))
	_, ok := lo.Find(plugins, func(plugin internal.Plugin) bool {
		return strings.HasPrefix(plugin.Url, golangCiLinter)
	})
	assert.True(suite.T(), ok)

	_, ok = lo.Find(plugins, func(plugin internal.Plugin) bool {
		return strings.HasPrefix(plugin.Url, testsum)
	})
	assert.True(suite.T(), ok)
	lo.ForEach(plugins, func(plugin internal.Plugin, _ int) {
		_, err := os.Stat(filepath.Join(suite.goPath, plugin.Binary()))
		assert.NoError(suite.T(), err)
	})
	_, err1 := os.Stat(filepath.Join(suite.testDir, "gob.yaml"))
	assert.NoError(suite.T(), err1)
}
