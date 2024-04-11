package cmd

import (
	"github.com/kcmvp/gob/gbc/artifact"
	"github.com/kcmvp/gob/utils"
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

type InitializeTestSuite struct {
	suite.Suite
}

func TestInitializeTestSuit(t *testing.T) {
	suite.Run(t, &InitializeTestSuite{})
}

func (suite *InitializeTestSuite) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))
}

func (suite *InitializeTestSuite) TestBuiltInPlugins() {
	plugins := builtinPlugins()
	assert.Equal(suite.T(), 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin artifact.Plugin) bool {
		return plugin.Module() == "github.com/golangci/golangci-lint"
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "golangci-lint", plugin.Name())
	assert.Equal(suite.T(), "lint", plugin.Alias)
}

func (suite *InitializeTestSuite) TestInitialization() {
	gopath := artifact.GoPath()
	target := artifact.CurProject().Target()
	initialize(nil, nil)
	plugins := artifact.CurProject().Plugins()
	assert.Equal(suite.T(), 2, len(plugins))
	_, ok := lo.Find(plugins, func(plugin artifact.Plugin) bool {
		return strings.HasPrefix(plugin.Url, golangCiLinter)
	})
	assert.True(suite.T(), ok)

	_, ok = lo.Find(plugins, func(plugin artifact.Plugin) bool {
		return strings.HasPrefix(plugin.Url, testsum)
	})
	assert.True(suite.T(), ok)
	lo.ForEach(plugins, func(plugin artifact.Plugin, _ int) {
		_, err := os.Stat(filepath.Join(gopath, plugin.Binary()))
		assert.NoError(suite.T(), err)
	})
	_, err1 := os.Stat(filepath.Join(target, "gob.yaml"))
	assert.NoError(suite.T(), err1)
	assert.Equal(suite.T(), []string{"build", "clean", "test", "lint"}, validBuilderArgs())
}
