package builder

import (
	"context"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gbt/builder/linter"
	"github.com/looplab/fsm"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type action string

var initialized action = "initialized"
var preCommitHook action = "pre-commit"
var commitMsgHook action = "commit-msg"
var prePushHook action = "pre-push"
var Clean action = "clean"
var Lint action = "linter"
var Test action = "test"
var Build action = "build"

//

const ctxCurrentBuilder = "_builder"

type Builder struct {
	fsm     *fsm.FSM
	project *Project
	repo    *git.Repository
	builtIn []action
	Report
	*buildOption
}

func NewBuilder(dir ...string) *Builder {
	var hook string
	var actions []action
	flag.StringVar(&hook, "hook", "", "git hook trigger")
	flag.Parse()
	// corner case: no hook parameter
	if len(hook) > 0 {
		log.Println(color.GreenString("triggered by %s", hook))
		actions = triggers(hook)
	}
	root := moduleDir(dir...)
	repo, err := git.PlainOpen(root)
	if err != nil {
		log.Println(color.YellowString("project is not at version control"))
	}
	return &Builder{
		fsm.NewFSM(string(initialized), events(), callBacks()),
		NewProject(root),
		repo,
		actions,
		Report{},
		defaultOption(),
	}
}

func moduleDir(dir ...string) string {
	if len(dir) > 0 {
		if _, err := os.ReadFile(filepath.Join(dir[0], "go.mod")); err == nil {
			return dir[0]
		}
	}
	_, file, _, ok := runtime.Caller(2)
	if ok {
		p := filepath.Dir(file)
		for p != string(os.PathSeparator) {
			if _, err := os.ReadFile(filepath.Join(p, "go.mod")); err == nil {
				return p
			} else {
				p = filepath.Dir(p)
			}
		}
	}
	panic(color.RedString("can't figure out project mod file"))
}

func triggers(hook string) []action {
	var target []action
	switch hook {
	case string(preCommitHook):
		target = append(target, preCommitHook, Lint)
	case string(commitMsgHook):
		target = append(target, commitMsgHook, Lint, Test)
	case string(prePushHook):
		target = append(target, prePushHook, Lint, Test)
	}
	return target
}

func events() fsm.Events {
	return fsm.Events{
		{string(Clean), []string{string(initialized)}, string(Clean)},
		// builtin for git commit
		{string(preCommitHook), []string{string(initialized)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(initialized)}, string(commitMsgHook)},
		{string(Clean), []string{string(initialized), string(commitMsgHook)}, string(Clean)},
		{string(Lint), []string{string(initialized), string(Clean), string(commitMsgHook), string(prePushHook), string(Test)}, string(Lint)},
		{string(Test), []string{string(initialized), string(Clean), string(commitMsgHook), string(prePushHook), string(Lint)}, string(Test)},
		{string(Build), []string{string(Lint), string(Test)}, string(Build)},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// Clean
		string(Clean): func(ctx context.Context, event *fsm.Event) {
			//GbtProject.clean()
			builder(ctx).project.clean()
		},
		// pre-commit
		string(preCommitHook): func(ctx context.Context, event *fsm.Event) {
			linter.Scan(builder(ctx).project.moduleDir, true)
		},
		// commit-msg
		string(commitMsgHook): func(ctx context.Context, event *fsm.Event) {
			//if msg, ok := ctx.Value(ctxCommitMsg).(string); ok {
			//	//hook.Validate(msg)
			//} else {
			//	event.Err = errors.New("missed commit message")
			//}
		},
		// pre-push
		string(prePushHook): func(ctx context.Context, event *fsm.Event) {
			//@todo
		},

		// Lint
		string(Lint): func(ctx context.Context, event *fsm.Event) {
			//linter.Scan(false)
			// 1: scan
			// 2: generate report only when all the data ready
		},
		fmt.Sprintf("after_%s", string(Lint)): func(ctx context.Context, event *fsm.Event) {
			// validate
		},
		// Test
		string(Test): func(ctx context.Context, event *fsm.Event) {
			// 1: test
			// 2: generate report only when all the data ready
			//GbtProject.test()
		},
		fmt.Sprintf("after_%s", string(Test)): func(ctx context.Context, event *fsm.Event) {
			// validate
		},
		// Build
		string(Build): func(ctx context.Context, event *fsm.Event) {
			//GbtProject.build()
		},
	}
}

func builder(ctx context.Context) *Builder {
	if v, ok := ctx.Value(ctxCurrentBuilder).(*Builder); ok {
		return v
	}
	panic(color.RedString("can't find the builder"))
}

func (builder *Builder) Run(actions ...action) {
	// 1: linter is mandatory
	// 2: setup internal action
	if len(actions) < 1 {
		log.Println(color.YellowString("please provide at least one action"))
		return
	}
	log.Printf("build actions are %+v \n", actions)
	ctx := context.WithValue(context.Background(), ctxCurrentBuilder, builder)
	for _, evt := range actions {
		err := builder.fsm.Event(ctx, string(evt))
		if err != nil {
			log.Fatalln(color.RedString("%v", err))
		}
	}
}
