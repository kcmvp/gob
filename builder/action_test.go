package builder

import (
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

type ActionTestSuite struct {
	suite.Suite
	project *Builder
	session *boot.Session
}

func (s *ActionTestSuite) SetupSuite() {
	s.project = NewBuilder()
}

func TestActionTestSuite(t *testing.T) {
	suite.Run(t, new(ActionTestSuite))
}

func (s *ActionTestSuite) SetupTest() {
	s.session = boot.NewSession()
}

func (s *ActionTestSuite) TestCleanAction() {
	lo.ForEach(mapper()[boot.Clean], func(action boot.Action, _ int) {
		err := action(s.session, s.project, boot.Clean)
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
	flags := lo.Filter(s.session.AllFlags(boot.Clean), func(k string, _ int) bool {
		return strings.HasPrefix(k, "clean.")
	})
	require.Empty(s.T(), flags, "should no flags")
}

/*
func (s *ActionTestSuite) TestBuildAction() {
	lo.ForEach(mapper()[boot.Build], func(action boot.Action, _ int) {
		err := action(s.session, s.project, boot.Build)
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
		err := action(s.session, s.project, boot.SetupHook)
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
*/
