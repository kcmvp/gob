package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/kcmvp/gob/builder"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CmdTestSuite struct {
	suite.Suite
	builder *builder.Builder
}

func (s *CmdTestSuite) SetupSuite() {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	for {
		if _, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
			os.Chdir(root)
			s.builder = builder.NewBuilder(root)
			break
		} else {
			root = filepath.Dir(root)
		}
	}
}

func TestCmdTestSuit(t *testing.T) {
	suite.Run(t, new(CmdTestSuite))
}

func (s *CmdTestSuite) TestSetupBuilder() {
	builder := filepath.Join(s.builder.ScriptDir(), "builder.go")
	_, err := os.Stat(builder)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", boot.SetupBuilder.Name()})
	err = rootCmd.Execute()
	if err == nil {
		require.NoError(s.T(), err, "should create builder.go successfully")
	} else {
		require.ErrorIs(s.T(), err, os.ErrExist, "should get file exists error")
	}
	require.Empty(s.T(), boot.AllFlags(boot.SetupBuilder))
}

func (s *CmdTestSuite) TestSetupHook() {

	for _, v := range builder.HookMap() {
		err := os.Remove(filepath.Join(s.builder.GitHome(), "hooks", v))
		require.True(s.T(), err == nil || errors.Is(err, os.ErrNotExist))
	}

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", boot.SetupHook.Name()})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)

	for k, v := range builder.HookMap() {
		f := filepath.Join(s.builder.ScriptDir(), fmt.Sprintf("%s.go", k))
		require.FileExists(s.T(), f)
		f = filepath.Join(s.builder.GitHome(), "hooks", v)
		require.FileExists(s.T(), f)
	}
	require.Empty(s.T(), boot.AllFlags(boot.SetupBuilder))

}

func (s *CmdTestSuite) TestSetupLint() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", boot.SetupLinter.Name(), "-v", "v1.49.0"})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.builder.Config().GetString("toolset.golangci-lint"), "v1.49.0")
	require.Equal(s.T(), "v1.49.0", boot.GetFlag[string](boot.SetupLinter, "version"))
}

func (s *CmdTestSuite) TestSetupLintWithLatest() {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"setup", boot.SetupLinter.Name()})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)
	require.Equal(s.T(), s.builder.Config().GetString("toolset.golangci-lint"), "v1.49.0")
	require.Equal(s.T(), boot.LatestVer, boot.GetFlag[string](boot.SetupLinter, "version"))
}

func (s *CmdTestSuite) TestRunLint() {
	tests := []struct {
		name   string
		flags  []string
		result bool
	}{
		{
			"changed",
			[]string{"run", boot.Lint.Name()},
			false,
		}, {
			"all",
			[]string{"run", boot.Lint.Name(), "-a"},
			true,
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			html := filepath.Join(s.builder.TargetDir(), "golangci-lint.html")
			out := filepath.Join(s.builder.TargetDir(), "golangci-lint.out")
			err := os.Remove(html)
			if err != nil {
				require.ErrorIs(s.T(), err, os.ErrNotExist)
			}
			err = os.Remove(out)
			if err != nil {
				require.ErrorIs(s.T(), err, os.ErrNotExist)
			}
			b := bytes.NewBufferString("")
			rootCmd.SetOut(b)
			rootCmd.SetArgs(test.flags)
			err = rootCmd.Execute()
			if err != nil {
				require.True(t, strings.Contains(err.Error(), "linter issues are found"))
			}
			require.Equal(s.T(), boot.GetFlag[bool](boot.Lint, "all"), test.result)
			require.Equal(t, boot.AllFlags(boot.Lint), []string{"all"})

			_, err = os.Stat(html)
			require.NoError(s.T(), err)
			_, err = os.Stat(out)
			require.NoError(s.T(), err)

		})
	}
}

func (s *CmdTestSuite) TestCleanWithCache() {
	tests := []struct {
		name     string
		flags    []string
		trueFlag string
		execCtx  string
	}{
		{
			"cleanCache",
			[]string{"run", boot.Clean.Name(), "--cache"},
			"-cache",
			"go clean -cache",
		},
		{
			"cleanCache_short",
			[]string{"run", boot.Clean.Name(), "-c"},
			"-cache",
			"go clean -cache",
		},
	}
	validFlags := []string{"-cache", "-testcache", "-modcache", "-fuzzcache"}
	sort.Strings(validFlags)
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			rootCmd.SetOut(b)
			rootCmd.SetArgs(test.flags)
			err := rootCmd.Execute()
			require.NoError(t, err)
			lo.ForEach([]string{"-cache", "-testcache", "-modcache", "-fuzzcache"}, func(flag string, _ int) {
				value := boot.GetFlag[bool](boot.Clean, flag)
				if flag == test.trueFlag {
					require.Equal(s.T(), true, value)
				} else {
					require.Equal(s.T(), false, value)
				}
			})
			expFlags := boot.AllFlags(boot.Clean)
			sort.Strings(expFlags)
			require.Equal(t, validFlags, expFlags)
			require.Equal(t, test.execCtx, boot.GetExecCtx(boot.Clean))
		})
	}
}

func (s *CmdTestSuite) TestCleanWithCacheAndMode() {
	tests := []struct {
		name      string
		flags     []string
		trueFlags []string
		execCtx   string
	}{
		{
			"cleanCache",
			[]string{"run", boot.Clean.Name(), "--cache", "--testcache"},
			[]string{"-cache", "-testcache"},
			"go clean -cache -testcache",
		},
		{
			"cleanCache_short",
			[]string{"run", boot.Clean.Name(), "-c", "-t"},
			[]string{"-cache", "-testcache"},
			"go clean -cache -testcache",
		},
	}
	validFlags := []string{"-cache", "-testcache", "-modcache", "-fuzzcache"}
	sort.Strings(validFlags)
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			rootCmd.SetOut(b)
			rootCmd.SetArgs(test.flags)
			err := rootCmd.Execute()
			require.NoError(t, err)
			lo.ForEach([]string{"-cache", "-testcache", "-modcache", "-fuzzcache"}, func(flag string, _ int) {
				value := boot.GetFlag[bool](boot.Clean, flag)
				if lo.Contains(test.trueFlags, flag) {
					require.Equal(s.T(), true, value)
				} else {
					require.Equal(s.T(), false, value)
				}
			})
			expFlags := boot.AllFlags(boot.Clean)
			sort.Strings(expFlags)
			require.Equal(t, validFlags, expFlags)
			require.Equal(t, test.execCtx, boot.GetExecCtx(boot.Clean))
		})
	}
}

func (s *CmdTestSuite) TestTestProject() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		return
	}
	os.Setenv("callFromTest", "1")
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"run", boot.Test.Name()})
	err := rootCmd.Execute()
	require.NoError(s.T(), err)

}
