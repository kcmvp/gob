package builder

import (
	"fmt"
	"github.com/kcmvp/gob/boot"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommandActionMapping(t *testing.T) {
	action1 := lo.Map(builderActions("pre_commit.go"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 := lo.Map([]boot.Action{cleanAction, lintAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok := lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("commit_msg.go"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{commitMsgAction, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("pre_push.go"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{cleanAction, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("builder"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, genBuilder}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("githook"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, getHook}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("linter"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, setupLinter}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("clean"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{cleanAction, getHook}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("lint"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, getHook, lintAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("test"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, getHook, testAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)

	action1 = lo.Map(builderActions("build"), func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	action2 = lo.Map([]boot.Action{createDirAction, getHook, buildAction}, func(t boot.Action, _ int) string {
		return fmt.Sprintf("%v", t)
	})
	ok = lo.Every(action1, action2)
	require.True(t, ok)
}
