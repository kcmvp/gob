package artifact

import (
	"fmt"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const depth = "github.com/KyleBanks/depth/cmd/depth"
const testsum = "gotest.tools/gotestsum"
const v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"

type ProjectTestSuite struct {
	suite.Suite
}

func (suite *ProjectTestSuite) BeforeTest(_, testName string) {
	s, _ := os.Open(filepath.Join(CurProject().Root(), "cmd", "gbc", "testdata", "gob.yaml"))
	root := filepath.Join(CurProject().Root(), "target", fmt.Sprintf("artifact_ProjectTestSuite_%s", testName))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "gob.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
}

func (suite *ProjectTestSuite) TearDownSuite() {
	_, method := utils.TestCaller()
	prefix := strings.TrimRight(method, "TearDownSuite")
	TearDownSuite(prefix)
}

func TestProjectSuite(t *testing.T) {
	suite.Run(t, &ProjectTestSuite{})
}

func TestBasic(t *testing.T) {
	plugins := CurProject().Plugins()
	assert.Equal(t, 0, len(plugins))
	assert.Equal(t, "github.com/kcmvp/gob", project.Module())
}

func (suite *ProjectTestSuite) TestPlugins() {
	_, err := os.Stat(filepath.Join(CurProject().Target(), "gob.yaml"))
	assert.NoError(suite.T(), err)
	plugins := CurProject().Plugins()
	fmt.Println(plugins)
	assert.Equal(suite.T(), 3, len(plugins))
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
		plugin, _ := NewPlugin(test.url)
		v := CurProject().isSetup(plugin)
		assert.Equal(suite.T(), test.settled, v)
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
	assert.True(suite.T(), lo.Contains(mainFiles, filepath.Join(CurProject().Root(), "cmd", "gbc", "gbc.go")))
}

func (suite *ProjectTestSuite) TestVersion() {
	assert.NotNil(suite.T(), Version())
}

func (suite *ProjectTestSuite) Test_Callee() {
	test, method := utils.TestCaller()
	assert.True(suite.T(), test)
	assert.Equal(suite.T(), "artifact_ProjectTestSuite_Test_Callee", method)
}

func (suite *ProjectTestSuite) TestHookDir() {
	hookDir := CurProject().HookDir()
	_, dir := utils.TestCaller()
	assert.True(suite.T(), strings.Contains(hookDir, dir))
	_, err := os.Stat(hookDir)
	assert.NoError(suite.T(), err)
}

func (suite *ProjectTestSuite) TestSetupPlugin() {
	plugin, _ := NewPlugin(v6)
	project.SetupPlugin(plugin)
	gopath := GoPath()
	entry, err := os.ReadDir(gopath)
	_, suffix := utils.TestCaller()
	assert.True(suite.T(), strings.HasSuffix(gopath, suffix))
	assert.NoErrorf(suite.T(), err, "GOPATH should be created")
	assert.True(suite.T(), len(entry) == 1, "plugin should be installed to GOPATH")
}
