package builder

import (
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type ActionTestSuite struct {
	suite.Suite
	project *Builder
}

func (s *ActionTestSuite) SetupSuite() {

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	for {
		if _, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
			os.Chdir(root)
			s.project = NewBuilder(root)
			break
		} else {
			root = filepath.Dir(root)
		}
	}
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

func (as *ActionTestSuite) TestHappyFlow() {
	require.Equal(as.T(), 1, 1)
}

func (s *ActionTestSuite) TestCleanAction() {
	lo.ForEach(mapper()[boot.Clean], func(action boot.Action, _ int) {
		err := action(s.project, boot.Clean)
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.project.TargetDir())
	/*
		f, err := os.Open(s.project.TargetDir())
		require.NoError(s.T(), err)
		defer f.Close()
		_, err = f.Readdirnames(1) // Or f.Readdir(1)
		require.ErrorIs(s.T(), err, io.EOF)
	*/
	flags := lo.Filter(boot.AllFlags(boot.Clean), func(k string, _ int) bool {
		return strings.HasPrefix(k, "clean.")
	})
	require.Empty(s.T(), flags, "should no flags")
}

func (s *ActionTestSuite) TestBuildAction() {
	lo.ForEach(mapper()[boot.Build], func(action boot.Action, _ int) {
		err := action(s.project, boot.Build)
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.project.TargetDir())
	found := false
	filepath.WalkDir(s.project.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if !found && !d.IsDir() {
			found = strings.HasPrefix(d.Name(), "main")
		}
		return err
	})
	require.True(s.T(), found, "file should exists")
}

func (s *ActionTestSuite) TestGitHookAction() {
	hm := lo.Invert(HookMap())
	filepath.WalkDir(filepath.Join(s.project.GitHome(), "hooks"), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if _, ok := hm[d.Name()]; ok {
				os.Remove(path)
			}
		}
		return err
	})

	lo.ForEach(mapper()[boot.SetupHook], func(action boot.Action, _ int) {
		err := action(s.project, boot.SetupHook)
		require.NoError(s.T(), err)
	})
	filepath.WalkDir(filepath.Join(s.project.GitHome(), "hooks"), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if _, ok := hm[d.Name()]; ok {
				hm[d.Name()] = "1"
			}
		}
		return err
	})
	for k, v := range hm {
		require.Equal(s.T(), "1", v, fmt.Sprintf("%s shold exists", k))
	}

}

/*
func (s *ActionTestSuite) TestLintFromHook() {

	s.project.in = "pre_commit.go"
	if c, ok := commandMap()["lint"]; ok {
		sort.Strings(c.Flags)
		f2 := []string{"-n", "--fix"}
		sort.Strings(f2)
		require.Equal(s.T(), c.Flags, f2)
		c.process(s.Context)

		sort.Strings(c.Flags)
		require.Equal(s.T(), c.Flags, f2)
	}
}
*/
