package builder

import (
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"
)

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
