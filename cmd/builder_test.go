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
}

func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, &BuilderTestSuit{})
}

func (suite *BuilderTestSuit) TearDownSuite() {
	TearDownSuite("cmd_builder_test_")
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
		err := builderCmd.Args(nil, test.args)
		assert.True(suite.T(), test.wantErr == (err != nil))
	}

}

func (suite *BuilderTestSuit) TestExecute() {
	builderCmd.SetArgs([]string{"cd"})
	err := Execute()
	assert.Error(suite.T(), err)
}

func (suite *BuilderTestSuit) TestBuild() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "invalid",
			args:    []string{"cd"},
			wantErr: true,
		},
		{
			name:    "valid",
			args:    []string{"build"},
			wantErr: false,
		},
	}
	for _, test := range tests {
		builderCmd.SetArgs(test.args)
		err := execute(builderCmd, test.args[0])
		assert.True(suite.T(), test.wantErr == (err != nil))
		if test.wantErr {
			assert.True(suite.T(), strings.Contains(err.Error(), color.RedString("")))
		}
	}
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

func (suite *BuilderTestSuit) TestRunE() {
	target := internal.CurProject().Target()
	err := builderCmd.RunE(builderCmd, []string{"build"})
	assert.NoError(suite.T(), err)
	_, err = os.Stat(filepath.Join(target, lo.If(internal.Windows(), "gob.exe").Else("gob")))
	assert.NoError(suite.T(), err, "binary should be generated")
	err = builderCmd.RunE(builderCmd, []string{"build", "clean"})
	assert.NoError(suite.T(), err)
	assert.NoFileExistsf(suite.T(), filepath.Join(target, lo.If(internal.Windows(), "gob.exe").Else("gob")), "binary should be deleted")
	err = builderCmd.RunE(builderCmd, []string{"def"})
	assert.Errorf(suite.T(), err, "can not find the command def")
}
