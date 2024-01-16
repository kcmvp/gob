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

type BuilderTestSuit struct {
	suite.Suite
	gopath  string
	testDir string
}

func TestBuilderTestSuit(t *testing.T) {
	_, file := internal.TestCallee()
	suite.Run(t, &BuilderTestSuit{
		gopath:  internal.GoPath(),
		testDir: filepath.Join(internal.CurProject().Target(), file),
	})
}

func (suite *BuilderTestSuit) SetupSuite() {
	os.RemoveAll(suite.gopath)
	os.RemoveAll(suite.testDir)
}

func (suite *BuilderTestSuit) TearDownTest() {
	os.RemoveAll(suite.gopath)
	os.RemoveAll(suite.testDir)
}

func (suite *BuilderTestSuit) TestValidArgs() {
	assert.Equal(suite.T(), []string{"build", "clean", "test", "lint"}, builderCmd.ValidArgs)
}

func (suite *BuilderTestSuit) TestArgs() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "not in valid args list",
			args:    []string{"def"},
			wantErr: true,
		},
		{
			name:    "partial valid args",
			args:    []string{"test", "def"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "empty args",
			args:    []string{""},
			wantErr: true,
		},
		{
			name:    "positive case",
			args:    []string{"clean", "test"},
			wantErr: false,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := builderCmd.Args(nil, test.args)
			assert.True(t, test.wantErr == (err != nil))
		})
	}

}

func (suite *BuilderTestSuit) TestBuild() {
	builderCmd.SetArgs([]string{"cd"})
	err := builderCmd.Execute()
	assert.Error(suite.T(), err)
	assert.True(suite.T(), strings.Contains(err.Error(), color.RedString("")))

}

func (suite *BuilderTestSuit) TestPersistentPreRun() {
	builderCmd.PersistentPreRun(nil, nil)
	hooks := lo.MapToSlice(internal.HookScripts(), func(key string, _ string) string {
		return key
	})
	for _, hook := range hooks {
		_, err := os.Stat(filepath.Join(internal.CurProject().HookDir(), hook))
		assert.Error(suite.T(), err)
	}
	internal.CurProject().SetupHooks(true)
	for _, hook := range hooks {
		_, err := os.Stat(filepath.Join(internal.CurProject().HookDir(), hook))
		assert.NoError(suite.T(), err)
	}
}

func (suite *BuilderTestSuit) TestBuiltinPlugins() {
	plugins := builtinPlugins()
	assert.Equal(suite.T(), 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin internal.Plugin) bool {
		return plugin.Url == "github.com/golangci/golangci-lint/cmd/golangci-lint"
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "v1.55.2", plugin.Version())
	assert.Equal(suite.T(), "golangci-lint", plugin.Name())
	assert.Equal(suite.T(), "github.com/golangci/golangci-lint", plugin.Module())
	assert.Equal(suite.T(), "lint", plugin.Alias)
	plugin, ok = lo.Find(plugins, func(plugin internal.Plugin) bool {
		return plugin.Url == "gotest.tools/gotestsum"
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "v1.11.0", plugin.Version())
	assert.Equal(suite.T(), "gotestsum", plugin.Name())
	assert.Equal(suite.T(), "gotest.tools/gotestsum", plugin.Module())
	assert.Equal(suite.T(), "test", plugin.Alias)
}
