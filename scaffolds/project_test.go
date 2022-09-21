package scaffolds

import (
	"encoding/json"
	"github.com/kcmvp/gob/boot"
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
	b.builder = NewProject()
}
func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

func (b *BuilderTestSuite) TestSetupBuilder() {
	boot.NewSession().Run(b.builder, boot.InitBuilder)
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
	boot.NewSession().Run(b.builder, boot.InitHook)
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
	boot.NewSession().Run(b.builder, boot.InitLinter)
	var last time.Time
	if err == nil {
		last = info.ModTime()
	}
	info, err = os.Stat(filepath.Join(b.builder.RootDir(), lintCfg))
	require.NoError(b.T(), err)
	require.True(b.T(), info.ModTime().UnixMilli() >= last.UnixMilli())
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
		session := boot.NewSession()
		if len(test.flag) > 0 {
			session.BindFlag(boot.Clean, test.flag, true)
		}
		session.Run(b.builder, boot.Clean)
		b.T().Run(test.name, func(t *testing.T) {
			require.Equal(t, test.ctxValue, session.CtxValue(boot.Clean))
		})
	}
}

func (b *BuilderTestSuite) TestLint() {
	tests := []struct {
		name     string
		scanAll  bool
		ctxValue string
	}{
		{
			"changesOnly",
			false,
			"run -v --out-format json ./... --new-from-rev HEAD~ golangci-lint-v1-49-0",
		},
		{
			"changesAll",
			true,
			"run -v --out-format json ./... --fix false golangci-lint-v1-49-0",
		},
	}
	for _, test := range tests {
		session := boot.NewSession()
		b.T().Run(test.name, func(t *testing.T) {
			session.BindFlag(boot.Lint, "all", test.scanAll)
			session.Run(b.builder, boot.Lint)
			require.Equal(t, test.ctxValue, session.CtxValue(boot.Lint))
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
	session := boot.NewSession()
	err := session.Run(b.builder, boot.Report)
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
	require.Equal(t, 3, len(mappers[boot.PreCommit]))
	require.Equal(t, 3, len(mappers[boot.CommitMsg]))
	require.Equal(t, 3, len(mappers[boot.PrePush]))
	require.Equal(t, 2, len(mappers[boot.InitBuilder]))
	require.Equal(t, 2, len(mappers[boot.InitHook]))
	require.Equal(t, 2, len(mappers[boot.InitLinter]))
	require.Equal(t, 2, len(mappers[boot.Clean]))
	require.Equal(t, 3, len(mappers[boot.Lint]))
	require.Equal(t, 3, len(mappers[boot.Test]))
	require.Equal(t, 3, len(mappers[boot.Build]))
	require.Equal(t, 5, len(mappers[boot.Report]))
}
