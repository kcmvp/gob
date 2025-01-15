package buildtime

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestWorkspace(t *testing.T) {
	root := RootDir()
	assert.NotNil(t, root)
	module := CurrentModule().MustGet()
	assert.Equal(t, "github.com/kcmvp/gob/buildtime", module.Path())
	assert.True(t, strings.HasPrefix(module.Dir(), root))
	modules := []string{"github.com/kcmvp/gob/buildtime", "github.com/kcmvp/gob/cmd/gob",
		"github.com/kcmvp/gob/common", "github.com/kcmvp/gob/dbo", "github.com/kcmvp/gob/runtime"}
	assert.ElementsMatch(t, modules, lo.Map(Modules(), func(m *Module, _ int) string {
		return m.Path()
	}))
}

func TestModule_MainFile(t *testing.T) {
	module := CurrentModule().MustGet()
	mf := module.MainFile()
	assert.True(t, mf.IsPresent())
	assert.True(t, strings.HasSuffix(mf.MustGet(), "/buildtime/dummy/dummy.go"))
}
