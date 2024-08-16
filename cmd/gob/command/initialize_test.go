package command

import (
	"encoding/json"
	"fmt"
	"github.com/kcmvp/gob/cmd/gob/artifact"
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

const golangCiLinter = "github.com/golangci/golangci-lint/cmd/golangci-lint"
const testsum = "gotest.tools/gotestsum"

type InitializeTestSuite struct {
	suite.Suite
}

func TestInitializeTestSuit(t *testing.T) {
	suite.Run(t, &InitializeTestSuite{})
}

func (suite *InitializeTestSuite) BeforeTest(_, testName string) {
	os.Chdir(artifact.CurProject().Root())
	var s, t *os.File
	s, _ = os.Open(filepath.Join(artifact.CurProject().Root(), "cmd", "gbc", "testdata", "config.json"))
	_, method := utils.TestCaller()
	root := filepath.Join(artifact.CurProject().Root(), "target", strings.ReplaceAll(method, "_BeforeTest", fmt.Sprintf("_%s", testName)))
	os.MkdirAll(root, os.ModePerm)
	t, _ = os.Create(filepath.Join(root, "config.json"))
	io.Copy(t, s)
	t.Close()
	s.Close()
}

func (suite *InitializeTestSuite) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))
}

func (suite *RootTestSuit) TestParsePlugins() {
	result, err := parseArtifacts(initializerCmd, []string{""}, "plugins")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.Exists())
	var plugins []artifact.Plugin
	err = json.Unmarshal([]byte(result.Raw), &plugins)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(plugins))
	assert.True(suite.T(), lo.ContainsBy(plugins, func(plugin artifact.Plugin) bool {
		return plugin.Alias == "lint"
	}))
}

func (suite *RootTestSuit) TestParseDeps() {
	result, err := parseArtifacts(initializerCmd, []string{""}, "deps")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.Exists())
	var deps []string
	err = json.Unmarshal([]byte(result.Raw), &deps)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(deps))
	assert.True(suite.T(), lo.Contains(deps, "github.com/stretchr/testify"))
}

func (suite *InitializeTestSuite) TestInitialization() {
	//gopath := artifact.GoPath()
	target := artifact.CurProject().Target()
	rootCmd.SetArgs([]string{"init"})
	rootCmd.Execute()
	//plugins := artifact.CurProject().Plugins()
	//assert.Equal(suite.T(), 3, len(plugins))
	//_, ok := lo.Find(plugins, func(plugin artifact.Plugin) bool {
	//	return strings.HasPrefix(plugin.Url, golangCiLinter)
	//})
	//assert.True(suite.T(), ok)
	//_, ok = lo.Find(plugins, func(plugin artifact.Plugin) bool {
	//	return strings.HasPrefix(plugin.Url, testsum)
	//})
	//assert.True(suite.T(), ok)
	//lo.ForEach(plugins, func(plugin artifact.Plugin, _ int) {
	//	_, err := os.Stat(filepath.Join(gopath, plugin.Binary()))
	//	assert.NoError(suite.T(), err)
	//})
	_, err1 := os.Stat(filepath.Join(target, "gob.yaml"))
	assert.NoError(suite.T(), err1)
	assert.True(suite.T(), lo.Every([]string{"build", "clean", "test", "depth", "lint"}, validBuilderArgs()))
}
