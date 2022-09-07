package builder

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"runtime"
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

/*
func (s *ActionTestSuite) TestCleanAction() {
	lo.ForEach(builderActions("clean"), func(action boot.Action, _ int) {
		err := action(s.DefaultProject, "clean")
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.DefaultProject.TargetDir())
	f, err := os.Open(s.DefaultProject.TargetDir())
	require.NoError(s.T(), err)
	defer f.Close()
	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	require.ErrorIs(s.T(), err, io.EOF)
	flags := lo.Filter(s.DefaultProject.AllKeys(), func(k string, _ int) bool {
		return strings.HasPrefix(k, "clean.")
	})
	require.Empty(s.T(), flags, "should no flags")
}

func (s *ActionTestSuite) TestBuildAction() {
	lo.ForEach(builderActions("build"), func(action boot.Action, _ int) {
		err := action(s.DefaultProject, "build")
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.DefaultProject.TargetDir())
	found := false
	filepath.WalkDir(s.DefaultProject.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if !found && !d.IsDir() {
			found = strings.HasPrefix(d.Name(), "main")
		}
		return err
	})
	require.True(s.T(), found, "file should exists")
}

func (s *ActionTestSuite) TestLintAction() {
	err := os.RemoveAll(filepath.Join(s.DefaultProject.TargetDir(), "golangci-lint.out"))
	require.NoError(s.T(), err)
	err = os.RemoveAll(filepath.Join(s.DefaultProject.TargetDir(), "golangci-lint.html"))
	require.NoError(s.T(), err)
	lo.ForEach(builderActions("lint"), func(action boot.Action, _ int) {
		err = action(s.DefaultProject, "lint")
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.DefaultProject.TargetDir())
	out := false
	html := false
	filepath.WalkDir(s.DefaultProject.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			out = out || d.Name() == "golangci-lint.out"
			html = html || d.Name() == "golangci-lint.html"
		}
		return err
	})
	require.True(s.T(), out, "golangci-lint.out should be generated")
	require.True(s.T(), html, "golangci-lint.html should be generated")
}

func (s *ActionTestSuite) TestGitHookAction() {
	lo.ForEach(builderActions("gitHook"), func(action boot.Action, _ int) {
		err := action(s.DefaultProject, "gitHook")
		require.NoError(s.T(), err)
	})
	require.DirExists(s.T(), s.DefaultProject.TargetDir())
	found := false
	filepath.WalkDir(s.DefaultProject.TargetDir(), func(path string, d fs.DirEntry, err error) error {
		if !found && !d.IsDir() {
			found = strings.HasPrefix(d.Name(), "main")
		}
		return err
	})
	require.True(s.T(), found, "file should exists")
}


func (s *ActionTestSuite) TestCleanActionExists() {
	err := os.MkdirAll(s.BuildWithCommand.TargetDir(), os.ModePerm)
	require.NoError(s.T(), err)
	if c, ok := commandMap()["clean"]; ok {
		_, err = c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.BuildWithCommand.TargetDir())
	} else {
		require.Fail(s.T(), "invalid command")
	}
}


func (s *ActionTestSuite) TestTestAction() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	if c, ok := commandMap()["test"]; ok {
		_, err := c.process(s.Context)
		require.NoError(s.T(), err)
		require.DirExists(s.T(), s.Builder.TargetDir())
	} else {
		require.Fail(s.T(), "builder is a valid command")
	}
}

func (s *ActionTestSuite) TestAlwaysGenerateHook() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		// fix infinite loop
		return
	}
	os.Setenv("callFromTest", "1")
	for _, a := range []string{"clean", "lint", "test", "build"} {
		for h, _ := range infra.Hooks() {
			err := os.Remove(filepath.Join(s.Builder.GitHome(), "hooks", h))
			require.NoError(s.T(), err)
		}
		for h, _ := range infra.Hooks() {
			_, err := os.Stat(filepath.Join(s.Builder.GitHome(), "hooks", h))
			require.Error(s.T(), err)
		}

		if c, ok := commandMap()[a]; ok {
			c.process(s.Context)
		}
		for h, _ := range infra.Hooks() {
			_, err := os.Stat(filepath.Join(s.Builder.GitHome(), "hooks", h))
			require.NoError(s.T(), err)
		}
	}
}

func (s *ActionTestSuite) TestLintFromHook() {

	s.Builder.hook = "pre_commit.go"
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
