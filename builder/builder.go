package builder

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gbt/builder/githook"
	"github.com/kcmvp/gbt/builder/linter"
	"github.com/kcmvp/gbt/builder/report"
	"github.com/looplab/fsm"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type action struct {
	name  string
	order int
}

var initialized = action{"initialized", 0}
var preCommitHook = action{"preCommitHook", 1}
var commitMsgHook = action{"commitMsgHook", 2}
var prePushHook = action{"prePushHook", 3}
var Clean = action{"clean", 4}
var Lint = action{"linter", 5}
var Test = action{"test", 6}
var Build = action{"build", 7}

//

type Builder struct {
	fsm     *fsm.FSM
	ctx     context.Context
	project *Project
	builtIn []action
	report.Quality
	*BuildOption
}

func NewBuilder() *Builder {
	return NewBuilderWith(DefaultOption())
}

func NewBuilderWith(option *BuildOption) *Builder {
	return &Builder{
		fsm.NewFSM(initialized.name, events(), callBacks()),
		context.Background(),
		NewProject(rootDir()),
		builtin(),
		report.Quality{},
		option,
	}
}

func rootDir() string {
	_, file, _, ok := runtime.Caller(3)
	if ok {
		p := filepath.Dir(file)
		for {
			if _, err := os.ReadFile(filepath.Join(p, "go.mod")); err == nil {
				return p
			} else {
				p = filepath.Dir(p)
			}
		}
	}
	panic("Can't figure out module directory")
}

func builtin() []action {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(0, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	frame, more := frames.Next()
	for more {
		for _, g := range githook.Hooks() {
			g = fmt.Sprintf("%s.go", g)
			if strings.HasSuffix(frame.File, g) {
				fmt.Printf("os.Args is %+v\n", os.Args)
				break
			}
		}
		frame, more = frames.Next()
	}
	return nil
}

func events() fsm.Events {
	return fsm.Events{
		{Clean.name, []string{initialized.name}, Clean.name},
		// builtin for git commit
		{preCommitHook.name, []string{initialized.name}, preCommitHook.name},
		{commitMsgHook.name, []string{initialized.name}, commitMsgHook.name},

		{Clean.name, []string{initialized.name, commitMsgHook.name}, Clean.name},

		{Lint.name, []string{initialized.name, Clean.name, commitMsgHook.name, prePushHook.name, Test.name}, Lint.name},
		{Test.name, []string{initialized.name, Clean.name, commitMsgHook.name, prePushHook.name, Lint.name}, Test.name},

		{Build.name, []string{Lint.name, Test.name}, Build.name},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// Clean
		Clean.name: func(ctx context.Context, event *fsm.Event) {
			//GbtProject.clean()
		},
		// pre-commit
		preCommitHook.name: func(ctx context.Context, event *fsm.Event) {
			linter.Scan(true)
		},
		// commit-msg
		commitMsgHook.name: func(ctx context.Context, event *fsm.Event) {
			//if msg, ok := ctx.Value(ctxCommitMsg).(string); ok {
			//	//hook.Validate(msg)
			//} else {
			//	event.Err = errors.New("missed commit message")
			//}
		},
		// pre-push
		prePushHook.name: func(ctx context.Context, event *fsm.Event) {
			//@todo
		},

		// Lint
		Lint.name: func(ctx context.Context, event *fsm.Event) {
			linter.Scan(false)
			// 1: scan
			// 2: generate report only when all the data ready
		},
		fmt.Sprintf("after_%s", Lint): func(ctx context.Context, event *fsm.Event) {
			// validate
		},
		// Test
		Test.name: func(ctx context.Context, event *fsm.Event) {
			// 1: test
			// 2: generate report only when all the data ready
			//GbtProject.test()
		},
		fmt.Sprintf("after_%s", Test): func(ctx context.Context, event *fsm.Event) {
			// validate
		},
		// Build
		Build.name: func(ctx context.Context, event *fsm.Event) {
			//GbtProject.build()
		},
	}
}

func (builder *Builder) Context() context.Context {
	return builder.ctx
}

func (builder *Builder) Run(actions ...action) {
	// 1: linter is mandatory
	// 2: setup internal action
	if len(actions) < 1 {
		log.Println(color.YellowString("please provide at least one action"))
		return
	}
	log.Printf("build actions are %+v \n", actions)
	for _, evt := range actions {
		err := builder.fsm.Event(builder.Context(), evt.name)
		if err != nil {
			log.Fatalln(color.RedString("%v", err))
		}
	}
}
