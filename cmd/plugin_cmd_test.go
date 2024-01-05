package cmd

import (
	"github.com/kcmvp/gob/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

var v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"
var v7 = "golang.org/x/tools/cmd/digraph@v0.16.1"

type PluginTestSuit struct {
	suite.Suite
	testDir string
	goPath  string
}

func (suite *PluginTestSuit) TearDownSuite() {
	os.RemoveAll(suite.goPath)
	os.RemoveAll(suite.testDir)
}

func TestPluginSuite(t *testing.T) {
	_, dir := internal.TestCallee()
	suite.Run(t, &PluginTestSuit{
		goPath:  internal.GoPath(),
		testDir: filepath.Join(internal.CurProject().Target(), dir),
	})
}

func (suite *PluginTestSuit) TestInstallPlugin() {
	install(nil, v6)
	plugins := internal.CurProject().Plugins()
	assert.Equal(suite.T(), 1, len(plugins))
	assert.Equal(suite.T(), "digraph", plugins[0].Name())
	assert.Equal(suite.T(), "v0.16.0", plugins[0].Version())
	assert.Equal(suite.T(), "golang.org/x/tools/cmd/digraph", plugins[0].Module())
	install(nil, v7)

	plugins = internal.CurProject().Plugins()
	assert.Equal(suite.T(), 1, len(plugins))
	_, err := os.Stat(filepath.Join(suite.goPath, plugins[0].Binary()))
	assert.NoError(suite.T(), err)
}

func (suite *PluginTestSuit) TestInstallPluginWithVersion() {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"no version", "github.com/hhatto/gocloc/cmd/gocloc", false},
		{"latest version", "github.com/hhatto/gocloc/cmd/gocloc@latest", false},
		{"incorrect version", "github.com/hhatto/gocloc/cmd/gocloc@abc", false},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := install(nil, test.url)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}

func (suite *PluginTestSuit) TestPluginArgs() {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			"no args",
			[]string{},
			true,
		},
		{
			"first not match",
			[]string{"def", "list"},
			true,
		},
		{
			"install without url",
			[]string{"install", ""},
			true,
		},
		{
			"install with url",
			[]string{"install", v6},
			false,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			err := pluginCmd.Args(nil, test.args)
			assert.True(t, test.wantErr == (err != nil))
		})
	}
}
