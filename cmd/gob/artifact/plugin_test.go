package artifact

import (
	"encoding/json"
	"fmt"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo" //nolint
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type InternalPluginTestSuit struct {
	suite.Suite
	//lintLatestVersion      string
	gotestsumLatestVersion string
}

func (suite *InternalPluginTestSuit) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.Trim(method, "TearDownSuite"))
}

func TestInternalPluginSuite(t *testing.T) {
	suite.Run(t, &InternalPluginTestSuit{
		gotestsumLatestVersion: LatestVersion(false, "gotest.tools/gotestsum")[0].B,
	})
}

func (suite *InternalPluginTestSuit) TestNewPlugin() {
	gopath := GoPath()
	defer func() {
		os.RemoveAll(gopath)
	}()
	tests := []struct {
		name    string
		url     string
		module  string
		logName string
		binary  string
		wantErr bool
	}{
		{
			name:    "specific version",
			url:     "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.1.1",
			module:  "github.com/golangci/golangci-lint",
			logName: "golangci-lint",
			binary:  "golangci-lint-v1.1.1",
			wantErr: false,
		},
		{
			name:    "has_@ but no version",
			url:     "github.com/golangci/golangci-lint/cmd/golangci-lint@",
			module:  "github.com/golangci/golangci-lint",
			logName: "",
			binary:  "-",
			wantErr: true,
		},
		{
			name:    "at the beginning of the url",
			url:     "@github.com/golangci/golangci-lint/cmd/golangci-lint",
			module:  "github.com/golangci/golangci-lint",
			logName: "",
			binary:  "-",
			wantErr: true,
		},
		{
			name:    "multiple@",
			url:     "github.com/golangci/golangci-lint/cmd/golangci@-lint@v1",
			module:  "github.com/golangci/golangci-lint",
			logName: "",
			binary:  "-",
			wantErr: true,
		},
		{
			name:    "gotestsum",
			url:     "gotest.tools/gotestsum",
			module:  "gotest.tools/gotestsum",
			logName: "gotestsum",
			binary:  fmt.Sprintf("%s-%s", "gotestsum", suite.gotestsumLatestVersion),
			wantErr: false,
		},
	}
	for _, test := range tests {
		suite.T().Run(test.name, func(t *testing.T) {
			plugin, err := NewPlugin(test.url)
			assert.Equal(t, plugin.taskName(), test.logName)
			assert.Equal(t, plugin.Binary(), test.binary)
			assert.True(t, test.wantErr == (err != nil))
			if !test.wantErr {
				assert.Equal(t, test.module, plugin.module)
			}
		})
	}
}

func (suite *InternalPluginTestSuit) TestUnmarshalJSON() {
	gopath := GoPath()
	defer func() {
		os.RemoveAll(gopath)
	}()
	data, _ := os.ReadFile(filepath.Join(CurProject().Root(), "cmd", "gob", "command", "resources", "config.json"))
	v := gjson.GetBytes(data, "gbc_init.plugins")
	var plugins []Plugin
	err := json.Unmarshal([]byte(v.Raw), &plugins)
	t := suite.T()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(plugins))
	// no command
	plugin, ok := lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == "gotest.tools/gotestsum"
	})
	assert.True(t, ok)
	assert.Equal(t, suite.gotestsumLatestVersion, plugin.Version())
	assert.Equal(t, "gotestsum", plugin.Name())
	assert.Equal(t, "gotest.tools/gotestsum", plugin.Module())
	assert.Equal(t, "test", plugin.Alias)
}

func (suite *InternalPluginTestSuit) TestInstallPlugin() {
	gopath := GoPath()
	defer func() {
		os.RemoveAll(gopath)
	}()
	t := suite.T()
	plugin, err := NewPlugin("gotest.tools/gotestsum")
	assert.NoError(t, err)
	path, err := plugin.install()
	assert.NoError(t, err)
	assert.NoFileExistsf(t, path, "temporay go path should be deleted")
	binary := filepath.Join(GoPath(), plugin.Binary())
	info1, err := os.Stat(binary)
	assert.NoErrorf(t, err, "testsum should be installed successfully")
	path, _ = plugin.install()
	assert.NoFileExistsf(t, path, "temporay go path should be deleted")
	info2, _ := os.Stat(binary)
	assert.Equal(t, info1.ModTime(), info2.ModTime())
}

func (suite *InternalPluginTestSuit) TestExecute() {
	gopath := GoPath()
	defer func() {
		os.RemoveAll(gopath)
	}()
	t := suite.T()
	plugin, err := NewPlugin("golang.org/x/tools/cmd/guru@v0.17.0")
	assert.NoError(t, err)
	err = plugin.Execute()
	fmt.Println(err.Error())
	assert.Error(t, err)
	//'exit status 2' means the plugin is executed but no parameters,
	assert.Equal(t, "exit status 2", err.Error())
	_, err = os.Stat(filepath.Join(CurProject().Target(), "start_guru.log"))
	assert.NoError(suite.T(), err)
}
