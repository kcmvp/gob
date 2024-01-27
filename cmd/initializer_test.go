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
}

func TestInitializationTestSuit(t *testing.T) {
	suite.Run(t, &InitializationTestSuite{})
}

func (suite *InitializationTestSuite) TearDownSuite() {
	TearDownSuite("cmd_initializer_test_")
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
	gopath := internal.GoPath()
	target := internal.CurProject().Target()
	initializerFunc(nil, nil)
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
		_, err := os.Stat(filepath.Join(gopath, plugin.Binary()))
		assert.NoError(suite.T(), err)
	})
	_, err1 := os.Stat(filepath.Join(target, "gob.yaml"))
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), []string{"build", "clean", "test", "lint"}, validBuilderArgs())
}
