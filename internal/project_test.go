package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

var v6 = "golang.org/x/tools/cmd/digraph@v0.16.0"
var v7 = "golang.org/x/tools/cmd/digraph@v0.16.1"

type ProjectTestSuit struct {
	suite.Suite
	goPath string
}

func TestProjectTestSuit(t *testing.T) {
	suite.Run(t, &ProjectTestSuit{
		goPath: GoPath(),
	})
}

func (suite *ProjectTestSuit) SetupTest() {
	os.RemoveAll(suite.goPath)
	println(suite.goPath)
}

func (suite *ProjectTestSuit) TestInstallPlugin() {
	CurProject().LoadSettings()
	cfg := CurProject().Configuration()
	defer func() {
		os.Remove(cfg)
	}()
	err := CurProject().InstallPlugin(v6, "callvis")
	assert.NoError(suite.T(), err)
	_, name := NormalizePlugin(v6)
	info, err := os.Stat(filepath.Join(suite.goPath, name))
	assert.NoError(suite.T(), err)
	println(info.Name())
	plugin, ok := lo.Find(CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v6
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "digraph", plugin.A)
	assert.Equal(suite.T(), "callvis", plugin.B)
	assert.Empty(suite.T(), plugin.C)
	assert.Equal(suite.T(), v6, plugin.D)

	// install same plugin again
	err = CurProject().InstallPlugin(v6, "callvis")
	assert.Error(suite.T(), err)

	// update the plugin
	err = CurProject().InstallPlugin(v7, "callvisv7", "run")
	assert.NoError(suite.T(), err)
	plugin, ok = lo.Find(CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == v7
	})
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 1, len(CurProject().Plugins()))
	assert.Equal(suite.T(), plugin.B, "callvisv7")
	assert.Equal(suite.T(), plugin.C, "run")
}

//func TestVersion(t *testing.T) {
//	assert.True(t, strings.HasPrefix(Version(), "v0.0.2"))
//	assert.True(t, strings.Contains(Version(), "@"))
//}
