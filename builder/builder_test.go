package builder

import (
	"encoding/json"
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type BuilderTestSuite struct {
	suite.Suite
	builder *Builder
}

func (b *BuilderTestSuite) SetupSuite() {
	b.builder = NewBuilder()
}
func TestBuilderTestSuit(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

func (b *BuilderTestSuite) TestSetupBuilder() {
	boot.NewSession().Run(b.builder, boot.SetupBuilder)
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
	boot.NewSession().Run(b.builder, boot.SetupHook)
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
	boot.NewSession().Run(b.builder, boot.SetupLinter)
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
	report := Report{}
	err = json.Unmarshal(data, &report)
	require.NoError(b.T(), err)
	require.True(b.T(), len(report.Pkgs) > 0)
}

func TestCommandActionMapping(t *testing.T) {
	action1 := lo.Map(mapper()[boot.PreCommit], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 := lo.Map([]boot.Action{cleanAction, lintAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok := lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.CommitMsg], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{commitMsgAction, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.PrePush], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{cleanAction, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.SetupBuilder], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genBuilder}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.SetupHook], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genHook}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.SetupLinter], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, setupLinter}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.Clean], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{cleanAction, genHook}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.Lint], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genHook, lintAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.Test], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genHook, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(mapper()[boot.Build], func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genHook, buildAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)
}
