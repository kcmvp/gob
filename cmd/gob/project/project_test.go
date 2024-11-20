package project

import (
	"fmt"
	"github.com/kcmvp/gob/cmd/gob/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type ProjectTestSuite struct {
	suite.Suite
}

func (suite *ProjectTestSuite) BeforeTest(_, testName string) {
	s, _ := os.Open(filepath.Join(RootDir(), "testdata", "build.yaml"))
	root := filepath.Join(RootDir(), "target", fmt.Sprintf("project_ProjectTestSuite_%s", testName))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "build.yaml"))
	io.Copy(t, s)
	s.Close()
	t.Close()
}

func (suite *ProjectTestSuite) TearDownSuite() {
	prefix := strings.TrimRight(utils.TestEnv().MustGet(), "TearDownSuite")
	TearDownSuite(prefix)
}

func TearDownSuite(prefix string) {
	filepath.WalkDir(os.TempDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
	filepath.WalkDir(filepath.Join(RootDir(), "target"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
}

func TestProjectSuite(t *testing.T) {
	suite.Run(t, &ProjectTestSuite{})
}

func (suite *ProjectTestSuite) TestBasic() {
	plugins := Plugins()
	assert.Len(suite.T(), plugins, 5)
	pluginName := lo.Map(plugins, func(item Plugin, _ int) string {
		return item.Name
	})
	assert.ElementsMatch(suite.T(), []string{"githook.pre-commit", "githook.pre-push", "githook.commit-msg", "lint", "test"}, pluginName)
	assert.Equal(suite.T(), "github.com/kcmvp/gob/cmd/gob", Module())
	assert.True(suite.T(), strings.HasSuffix(TargetDir(), "gob/target/project_ProjectTestSuite_TestBasic"))
	assert.True(suite.T(), strings.HasSuffix(RootDir(), "gob"))
	assert.True(suite.T(), lo.ContainsBy(MainFiles(), func(item string) bool {
		return strings.HasSuffix(item, "gob/cmd/gob/gob.go")
	}))
	assert.True(suite.T(), strings.HasSuffix(CacheDir(), filepath.Join(".gob", Module())))
	hookDor := hookDir("githook")
	assert.True(suite.T(), strings.HasSuffix(hookDor, `git/hooks`))
	assert.NotEmpty(suite.T(), GoPath())
	assert.NotEmpty(suite.T(), temporaryGoPath())

}

func TestDeps(t *testing.T) {
	deps := Dependencies()
	assert.Equal(t, 12, len(deps))
	dep := mo.TupleToOption(lo.Find(deps, func(dep Dependency) bool {
		return dep.module == "github.com/fatih/color"
	}))
	assert.True(t, dep.IsPresent())
	assert.Len(t, dep.MustGet().dependencies, 3)
}
