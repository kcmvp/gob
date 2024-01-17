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

const depth = "github.com/KyleBanks/depth/cmd/depth"
const testsum = "gotest.tools/gotestsum"

type ProjectTestSuite struct {
	suite.Suite
	goPath  string
	testDir string
}

func (suite *ProjectTestSuite) SetupSuite() {
	s, _ := os.Open(filepath.Join(CurProject().Root(), "testdata", "gob.yaml"))
	os.MkdirAll(suite.testDir, os.ModePerm)
	t, _ := os.Create(filepath.Join(suite.testDir, "gob.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
}

func (suite *ProjectTestSuite) TearDownSuite() {
	//os.RemoveAll(suite.testDir)
	os.RemoveAll(suite.goPath)
}

func TestProjectSuite(t *testing.T) {
	_, file := TestCallee()
	suite.Run(t, &ProjectTestSuite{
		goPath:  GoPath(),
		testDir: filepath.Join(CurProject().Target(), file),
	})
}

func (suite *ProjectTestSuite) TestPlugins() {
	_, err := os.Stat(filepath.Join(suite.testDir, "gob.yaml"))
	assert.NoError(suite.T(), err)
	plugins := CurProject().Plugins()
	fmt.Println(plugins)
	assert.Equal(suite.T(), 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == testsum
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), testsum, plugin.Module())
	assert.Equal(suite.T(), "v1.11.0", plugin.Version())
	assert.Equal(suite.T(), "gotestsum", plugin.Name())
	assert.Equal(suite.T(), "test", plugin.Alias)
	plugin, ok = lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == depth
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "github.com/KyleBanks/depth", plugin.Module())
	assert.Equal(suite.T(), "v1.2.1", plugin.Version())
	assert.Equal(suite.T(), "depth", plugin.Name())
	assert.Equal(suite.T(), "depth", plugin.Alias)
}

func (suite *ProjectTestSuite) TestIsSetup() {
	tests := []struct {
		name    string
		url     string
		settled bool
	}{
		{
			name:    "no version",
			url:     depth,
			settled: true,
		},
		{
			name:    "with version",
			url:     fmt.Sprintf("%s@latest", depth),
			settled: true,
		},
		{
			name:    "specified version",
			url:     fmt.Sprintf("%s@v1.1.1", depth),
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
	CurProject().GitHook()
	for name, _ := range HookScripts() {
		_, err := os.Stat(filepath.Join(CurProject().HookDir(), name))
		assert.NoError(suite.T(), err)
	}
}

func (suite *ProjectTestSuite) TestMainFiles() {
	mainFiles := CurProject().MainFiles()
	assert.Equal(suite.T(), 1, len(mainFiles))
	assert.True(suite.T(), lo.Contains(mainFiles, filepath.Join(CurProject().Root(), "gob.go")))
}
