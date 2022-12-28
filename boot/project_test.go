package boot

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type BuilderTestSuite struct {
	suite.Suite
	builder *Project
}

func (b *BuilderTestSuite) SetupSuite() {
	pwd, _ := os.Getwd()
	b.builder = NewProject(pwd)
}
func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

func (b *BuilderTestSuite) TestSetupBuilder() {
	NewSession().Run(b.builder, SetupBuilder)
	_, err := os.Stat(filepath.Join(b.builder.ScriptDir(), "builder.go"))
	require.NoError(b.T(), err)
}

func (b *BuilderTestSuite) TestSetupHook() {
	err := os.Remove(filepath.Join(b.builder.GitHome(), "hooks", "pre-push"))
	require.NoError(b.T(), err)
	err = os.Remove(filepath.Join(b.builder.GitHome(), "hooks", "pre-commit"))
	require.NoError(b.T(), err)
	err = os.Remove(filepath.Join(b.builder.GitHome(), "hooks", "commit-msg"))
	require.NoError(b.T(), err)
	NewSession().Run(b.builder, SetupHook)
	_, err = os.Stat(filepath.Join(b.builder.ScriptDir(), "pre_commit.go"))
	require.NoError(b.T(), err)
	_, err = os.Stat(filepath.Join(b.builder.ScriptDir(), "pre_push.go"))
	require.NoError(b.T(), err)
	_, err = os.Stat(filepath.Join(b.builder.ScriptDir(), "commit_msg.go"))
	require.NoError(b.T(), err)
	_, err = os.Stat(filepath.Join(b.builder.GitHome(), "hooks", "pre-push"))
	require.NoError(b.T(), err)
	_, err = os.Stat(filepath.Join(b.builder.GitHome(), "hooks", "commit-msg"))
	require.NoError(b.T(), err)
	_, err = os.Stat(filepath.Join(b.builder.GitHome(), "hooks", "pre-commit"))
	require.NoError(b.T(), err)
}

func (b *BuilderTestSuite) TestSetupLinter() {
	info, err := os.Stat(filepath.Join(b.builder.RootDir(), lintCfg))
	NewSession().Run(b.builder, SetupLinter)
	var last time.Time
	if err == nil {
		last = info.ModTime()
	}
	info, err = os.Stat(filepath.Join(b.builder.RootDir(), lintCfg))
	require.NoError(b.T(), err)
	require.True(b.T(), info.ModTime().UnixMilli() >= last.UnixMilli())
}

func (b *BuilderTestSuite) TestSetupGitFlow() {
	NewSession().Run(b.builder, SetupGitFlow)
	//var last time.Time
	//if err == nil {
	//	last = info.ModTime()
	//}
	//info, err = os.Stat(filepath.Join(b.builder.RootDir(), lintCfg))
	//require.NoError(b.T(), err)
	//require.True(b.T(), info.ModTime().UnixMilli() >= last.UnixMilli())
}

func (b *BuilderTestSuite) TestClean() {
	tests := []struct {
		name     string
		flag     string
		ctxValue string
		empty    bool
	}{
		{
			"normal",
			"",
			"go clean delete=false",
			false,
		},
		{
			"clean cache",
			"-testcache",
			"go clean -testcache delete=false",
			false,
		},
	}
	for _, test := range tests {
		session := NewSession()
		if len(test.flag) > 0 {
			session.BindFlag(Clean, test.flag, true)
		}
		session.Run(b.builder, Clean)
		b.T().Run(test.name, func(t *testing.T) {
			require.Equal(t, test.ctxValue, session.CtxValue(Clean))
		})
	}
}

func (b *BuilderTestSuite) TestLint() {
	tests := []struct {
		name     string
		scanAll  bool
		ctxValue string
	}{
		//{
		//	"changesOnly",
		//	false,
		//	"run -v --out-format json ./... --new-from-rev HEAD~ golangci-lint-v1-49-0",
		//},
		{
			"changesAll",
			true,
			"run -v --out-format json ./... --fix false golangci-lint-v1-49-0",
		},
	}
	for _, test := range tests {
		session := NewSession()
		b.T().Run(test.name, func(t *testing.T) {
			session.BindFlag(Lint, "all", test.scanAll)
			session.Run(b.builder, Lint)
			require.Equal(t, test.ctxValue, session.CtxValue(Lint))
			_, err := os.Stat(filepath.Join(b.builder.TargetDir(), "lint.html"))
			require.NoError(t, err)
			_, err = os.Stat(filepath.Join(b.builder.TargetDir(), "lint.out"))
			require.NoError(t, err)
			_, err = os.Stat(filepath.Join(b.builder.TargetDir(), "lint.html"))
			require.NoError(t, err)
		})
	}
}

func (b *BuilderTestSuite) TestReportCommand() {
	if _, ok := os.LookupEnv("callFromTest"); ok {
		return
	}
	os.Setenv("callFromTest", "1")
	session := NewSession()
	err := session.Run(b.builder, Report)
	require.NoError(b.T(), err)
	data, err := os.ReadFile(filepath.Join(b.builder.TargetDir(), reportJSON))
	require.NoError(b.T(), err)
	report := BuildReport{}
	err = json.Unmarshal(data, &report)
	require.NoError(b.T(), err)
	require.True(b.T(), len(report.Pkgs) > 0)
}

func TestCommandActionMapping(t *testing.T) {
	mappers := mapper()
	require.Equal(t, 3, len(mappers[PreCommit]))
	require.Equal(t, 3, len(mappers[CommitMsg]))
	require.Equal(t, 3, len(mappers[PrePush]))
	require.Equal(t, 3, len(mappers[SetupBuilder]))
	require.Equal(t, 3, len(mappers[SetupHook]))
	require.Equal(t, 2, len(mappers[SetupLinter]))
	require.Equal(t, 2, len(mappers[Clean]))
	require.Equal(t, 3, len(mappers[Lint]))
	require.Equal(t, 3, len(mappers[Test]))
	require.Equal(t, 4, len(mappers[Build]))
	require.Equal(t, 5, len(mappers[Report]))
}
