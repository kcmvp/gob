package command

import (
	"github.com/kcmvp/gob/cmd/gob/artifact"
	"github.com/kcmvp/gob/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"
const v7 = "golang.org/x/tools/cmd/digraph@v0.16.1"

type PluginTestSuit struct {
	suite.Suite
}

func TestPluginSuite(t *testing.T) {
	suite.Run(t, &PluginTestSuit{})
}

func (suite *PluginTestSuit) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))

}
func (suite *PluginTestSuit) TestInstallPlugin() {
	err := install(nil, v6)
	assert.NoError(suite.T(), err)
	plugins := artifact.CurProject().Plugins()
	_, err = os.Stat(filepath.Join(artifact.GoPath(), plugins[0].Binary()))
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(plugins))
	assert.Equal(suite.T(), "digraph", plugins[0].Name())
	assert.Equal(suite.T(), "v0.16.0", plugins[0].Version())
	assert.Equal(suite.T(), "golang.org/x/tools/cmd/digraph", plugins[0].Module())
	err = install(nil, v7)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(plugins))
	_, err = os.Stat(filepath.Join(artifact.GoPath(), plugins[0].Binary()))
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
		{"incorrect version", "github.com/hhatto/gocloc/cmd/gocloc@abc", true},
	}
	for _, test := range tests {
		err := install(nil, test.url)
		assert.True(suite.T(), test.wantErr == (err != nil))
	}
}

func (suite *PluginTestSuit) TestList() {
	err := list(nil, "")
	assert.NoError(suite.T(), err)
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
		err := pluginCmd.Args(nil, test.args)
		assert.True(suite.T(), test.wantErr == (err != nil))
	}
}

func (suite *PluginTestSuit) TestRunE() {
	err := pluginCmd.RunE(pluginCmd, []string{"list"})
	assert.NoErrorf(suite.T(), err, "should list iinstalled plugin successfully")
}

func (suite *PluginTestSuit) TestPluginHelpTemplate() {
	rootCmd.SetArgs([]string{"plugin", "--help"})
	rootCmd.Execute()
}
