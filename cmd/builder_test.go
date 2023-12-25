package cmd

import (
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type BuilderTestSuit struct {
	suite.Suite
	start  int64
	gopath string
}

func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, &BuilderTestSuit{
		start:  time.Now().UnixNano(),
		gopath: os.Getenv("GOPATH"),
	})
}

func (suite *BuilderTestSuit) TestPersistentPreRun() {
	preRun()
	hooks := lo.MapToSlice(internal.HookScripts, func(key string, _ string) string {
		return key
	})
	for _, hook := range hooks {
		info, err := os.Stat(filepath.Join(internal.CurProject().HookDir(), hook))
		if err == nil {
			assert.True(suite.T(), info.ModTime().UnixNano() > suite.start)
		}
	}
	// test the missing plugins installation
	lo.ForEach(internal.CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string], index int) {
		_, name := internal.NormalizePlugin(plugin.D)
		_, err := os.Stat(filepath.Join(suite.gopath, "bin", name))
		assert.NoErrorf(suite.T(), err, "plugin should be insalled")
	})
}
