package internal

import (
	"encoding/json"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"os"
	"path/filepath"
	"testing"
)

func TestNewPlugin(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		module  string
		wantErr bool
	}{
		{"without version",
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/golangci/golangci-lint",
			false,
		},
		{"laatest version",
			"github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
			"github.com/golangci/golangci-lint",
			false,
		},
		{"specific version",
			"github.com/golangci/golangci-lint/cmd/golangci-lint@v1.1.1",
			"github.com/golangci/golangci-lint",
			false,
		},
		{"has @ but no version",
			"github.com/golangci/golangci-lint/cmd/golangci-lint@",
			"github.com/golangci/golangci-lint",
			true,
		},
		{"at the beginning of the url",
			"@github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/golangci/golangci-lint",
			true,
		},
		{"multiple @",
			"github.com/golangci/golangci-lint/cmd/golangci@-lint@v1",
			"github.com/golangci/golangci-lint",
			true,
		},

		// gotest.tools/gotestsum
		{"gotestsum",
			"gotest.tools/gotestsum",
			"gotest.tools/gotestsum",
			false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin, err := NewPlugin(test.url)
			assert.True(t, test.wantErr == (err != nil))
			if !test.wantErr {
				assert.Equal(t, test.module, plugin.module)
				assert.True(t, lo.Contains([]string{"v1.55.2", "v1.1.1", "v1.11.0"}, plugin.Version()))
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	data, _ := os.ReadFile(filepath.Join(CurProject().Root(), "cmd", "resources", "config.json"))
	v := gjson.GetBytes(data, "plugins")
	var plugins []Plugin
	err := json.Unmarshal([]byte(v.Raw), &plugins)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(plugins))
	plugin, ok := lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == "github.com/golangci/golangci-lint/cmd/golangci-lint"
	})
	assert.True(t, ok)
	assert.Equal(t, "v1.55.2", plugin.Version())
	assert.Equal(t, "golangci-lint", plugin.Name())
	assert.Equal(t, "github.com/golangci/golangci-lint", plugin.Module())
	assert.Equal(t, "lint", plugin.Alias)
	// no command
	plugin, ok = lo.Find(plugins, func(plugin Plugin) bool {
		return plugin.Url == "gotest.tools/gotestsum"
	})
	assert.True(t, ok)
	assert.Equal(t, "v1.11.0", plugin.Version())
	assert.Equal(t, "gotestsum", plugin.Name())
	assert.Equal(t, "gotest.tools/gotestsum", plugin.Module())
	assert.Equal(t, "test", plugin.Alias)

}
