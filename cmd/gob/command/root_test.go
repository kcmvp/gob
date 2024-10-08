package command

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/artifact"
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type RootTestSuit struct {
	suite.Suite
}

func TestRootTestSuit(t *testing.T) {
	suite.Run(t, &RootTestSuit{})
}

func (suite *RootTestSuit) BeforeTest(_, testName string) {
	os.Chdir(artifact.CurProject().Root())
	s, _ := os.Open(filepath.Join(artifact.CurProject().Root(), "cmd", "gob", "testdata", "config.json"))
	_, method := utils.TestCaller()
	root := filepath.Join(artifact.CurProject().Root(), "target", strings.ReplaceAll(method, "_BeforeTest", fmt.Sprintf("_%s", testName)))
	os.MkdirAll(root, os.ModePerm)
	t, _ := os.Create(filepath.Join(root, "config.json"))
	io.Copy(t, s)
	t.Close()
	s.Close()
}

func (suite *RootTestSuit) TearDownSuite() {
	_, method := utils.TestCaller()
	TearDownSuite(strings.TrimRight(method, "TearDownSuite"))
}

func (suite *RootTestSuit) TestValidArgs() {
	args := rootCmd.ValidArgs
	assert.Equal(suite.T(), 4, len(args))
	assert.True(suite.T(), lo.Every(args, []string{"build", "clean", "test", "lint"}))
}

func (suite *RootTestSuit) TestArgs() {
	artifact.CurProject().SetupHooks(true)
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "not in valid args list",
			args:    []string{"def"},
			wantErr: true,
		},
		{
			name:    "partial valid args",
			args:    []string{"build", "def"},
			wantErr: true,
		},
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "positive case",
			args:    []string{"clean", "build"},
			wantErr: false,
		},
	}
	for _, test := range tests {
		rootCmd.SetArgs(test.args)
		err := Execute()
		assert.True(suite.T(), test.wantErr == (err != nil))
	}

}

func (suite *RootTestSuit) TestExecute() {
	artifact.CurProject().SetupHooks(true)
	os.Chdir(artifact.CurProject().Target())
	rootCmd.SetArgs([]string{"build"})
	err := Execute()
	assert.Equal(suite.T(), "Please execute the command in the project root dir", err.Error())
	rootCmd.SetArgs([]string{"cd"})
	err = Execute()
	lo.ForEach([]string{"build", "clean", "test", "lint", "depth"}, func(item string, _ int) {
		assert.Equal(suite.T(), "invalid argument \"cd\" for gbc", err.Error())
	})
	os.Chdir(artifact.CurProject().Root())
	rootCmd.SetArgs([]string{"build"})
	err = Execute()
	assert.NoError(suite.T(), err)
}

func (suite *RootTestSuit) TestBuild() {
	artifact.CurProject().SetupHooks(true)
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args",
			wantErr: true,
		},
		{
			name:    "invalid",
			args:    []string{"cd"},
			wantErr: true,
		},
		{
			name:    "valid",
			args:    []string{"build"},
			wantErr: false,
		},
	}
	for _, test := range tests {
		rootCmd.SetArgs(test.args)
		err := rootCmd.Execute()
		assert.True(suite.T(), test.wantErr == (err != nil))
		if test.wantErr {
			assert.True(suite.T(), strings.Contains(err.Error(), color.RedString("")))
		}
	}
}

func (suite *RootTestSuit) TestPersistentPreRun() {
	artifact.CurProject().SetupHooks(true)
	rootCmd.SetArgs([]string{"build"})
	Execute()
	hooks := lo.MapToSlice(artifact.HookScripts(), func(key string, _ string) string {
		return key
	})
	for _, hook := range hooks {
		_, err := os.Stat(filepath.Join(artifact.CurProject().HookDir(), hook))
		assert.NoError(suite.T(), err)
	}
}

func (suite *RootTestSuit) TestRunE() {
	target := artifact.CurProject().Target()
	err := rootCmd.RunE(rootCmd, []string{"build"})
	assert.NoError(suite.T(), err)
	_, err = os.Stat(filepath.Join(target, lo.If(artifact.Windows(), "gob.exe").Else("gob")))
	assert.NoError(suite.T(), err, "binary should be generated")
	err = rootCmd.RunE(rootCmd, []string{"build", "clean"})
	assert.NoError(suite.T(), err)
	assert.NoFileExistsf(suite.T(), filepath.Join(target, lo.If(artifact.Windows(), "gob.exe").Else("gob")), "binary should be deleted")
	err = rootCmd.RunE(rootCmd, []string{"def"})
	assert.Errorf(suite.T(), err, "can not find the command def")
}

func (suite *RootTestSuit) TestOutOfRoot() {
	os.Chdir(artifact.CurProject().Target())
	err := Execute()
	assert.Error(suite.T(), err)
	assert.True(suite.T(), strings.Contains(err.Error(), "Please execute the command in the project root dir"))
}

func TearDownSuite(prefix string) {
	filepath.WalkDir(os.TempDir(), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
	filepath.WalkDir(filepath.Join(artifact.CurProject().Root(), "target"), func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
			os.RemoveAll(path)
		}
		return nil
	})
}
