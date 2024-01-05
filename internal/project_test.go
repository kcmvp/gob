package internal

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const golanglinturl = "github.com/golangci/golangci-lint/cmd/golangci-lint"

type ProjectTestSuite struct {
	suite.Suite
	goPath  string
	testDir string
}

func (suite *ProjectTestSuite) SetupSuite() {
	s, _ := os.Open(filepath.Join(CurProject().Root(), "gob.yaml"))
	os.MkdirAll(suite.testDir, os.ModePerm)
	t, _ := os.Create(filepath.Join(suite.testDir, "gob.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
	CurProject().LoadSettings()
}

func (suite *ProjectTestSuite) TearDownSuite() {
	os.RemoveAll(suite.testDir)
	os.RemoveAll(suite.goPath)
}

func TestSuite(t *testing.T) {
	_, file := TestCallee()
	CurProject().LoadSettings()
	suite.Run(t, &ProjectTestSuite{
		goPath:  GoPath(),
		testDir: filepath.Join(CurProject().Target(), file),
	})
}

func (suite *ProjectTestSuite) TestPlugins() {
	plugins := CurProject().Plugins()
	assert.Equal(suite.T(), 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == "gotest.tools/gotestsum"
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "gotest.tools/gotestsum", plugin.Module())
	assert.Equal(suite.T(), "v1.11.0", plugin.Version())
	assert.Equal(suite.T(), "gotestsum", plugin.Name())
	assert.Equal(suite.T(), "test", plugin.Alias)
	plugin, ok = lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == golanglinturl
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "github.com/golangci/golangci-lint", plugin.Module())
	assert.Equal(suite.T(), "v1.55.2", plugin.Version())
	assert.Equal(suite.T(), "golangci-lint", plugin.Name())
	assert.Equal(suite.T(), "lint", plugin.Alias)
}

func (suite *ProjectTestSuite) TestIsSetup() {
	tests := []struct {
		name    string
		url     string
		settled bool
	}{
		{
			name:    "no version",
			url:     golanglinturl,
			settled: true,
		},
		{
			name:    "with version",
			url:     fmt.Sprintf("%s@latest", golanglinturl),
			settled: true,
		},
		{
			name:    "specified version",
			url:     fmt.Sprintf("%s@v1.1.1", golanglinturl),
			settled: true,
		},
		{
			name:    "no settled",
			url:     "entgo.io/ent/cmd/ent",
			settled: false,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(j *testing.T) {
			plugin, _ := NewPlugin(test.url)
			v := CurProject().isSetup(plugin)
			assert.Equal(j, test.settled, v)
		})
	}
}

func (suite *ProjectTestSuite) TestValidate() {
	CurProject().Validate()
	lo.ForEach(CurProject().Plugins(), func(plugin Plugin, _ int) {
		info, err := os.Stat(filepath.Join(suite.goPath, plugin.Binary()))
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), info.Size() > 0)
	})
	CurProject().GitHook()
	for name, _ := range HookScripts() {
		_, err := os.Stat(filepath.Join(CurProject().HookDir(), name))
		assert.NoError(suite.T(), err)
	}
}
