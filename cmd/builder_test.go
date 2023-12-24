package cmd

import (
	"bytes"
	"github.com/kcmvp/gb/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var hookDir = filepath.Join(internal.CurProject().Root(), ".git", "hooks")

type BuilderTestSuit struct {
	suite.Suite
	start  int64
	gopath string
}

func (suite *BuilderTestSuit) SetupSuite() {
	hooks := lo.MapToSlice(internal.HookScripts, func(key string, _ string) string {
		return key
	})
	filepath.WalkDir(hookDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if lo.Contains(hooks, d.Name()) {
			os.Remove(path)
		}
		return nil
	})
}

func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, &BuilderTestSuit{
		start:  time.Now().UnixNano(),
		gopath: os.Getenv("GOPATH"),
	})
}

func (suite *BuilderTestSuit) TestValidArgs() {
	assert.Equal(suite.T(), 4, len(builderCmd.ValidArgs))
	assert.True(suite.T(), lo.Every(builderCmd.ValidArgs, []string{"build", "clean", "test", "lint"}))
}

func (suite *BuilderTestSuit) TestPersistentPreRun() {
	b := bytes.NewBufferString("")
	builderCmd.SetOut(b)
	builderCmd.PersistentPreRun(nil, nil)
	hooks := lo.MapToSlice(internal.HookScripts, func(key string, _ string) string {
		return key
	})
	for _, hook := range hooks {
		info, err := os.Stat(filepath.Join(hookDir, hook))
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), info.ModTime().UnixNano() > suite.start)
	}
	// test the missing plhgins installation
	lo.ForEach(internal.CurProject().Plugins(), func(plugin lo.Tuple4[string, string, string, string], index int) {
		_, name := internal.NormalizePlugin(plugin.D)
		_, err := os.Stat(filepath.Join(suite.gopath, "bin", name))
		assert.NoErrorf(suite.T(), err, "plugin should be insalled")
	})
}
