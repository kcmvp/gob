package builder

import (
	"context"
	"log"

	"github.com/kcmvp/gob/infra"
	"github.com/looplab/fsm"
)

var genBuilderCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	log.Println("Creating project build file")
	err := infra.SetupBuilder(GetBuilder(ctx).ScriptDir())
	if err != nil {
		event.Cancel(err)
	}
}

var genGitHookCallback fsm.Callback = func(ctx context.Context, event *fsm.Event) {
	value, ok := ctx.Value(GenHook).(bool)
	builder := GetBuilder(ctx)
	gitHome, err := builder.GitHome()
	cancelEvent(event, err)
	if err != nil {
		event.Cancel(err)
	}
	err = infra.GenGitHooks(gitHome, builder.ScriptDir(), ok && value)
	cancelEvent(event, err, "failed to setup hook")
}
