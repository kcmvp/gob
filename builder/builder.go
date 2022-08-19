package builder

import (
	"context"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gbt/builder/githook"
	"github.com/kcmvp/gbt/builder/linter"
	"github.com/looplab/fsm"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type action string

var initialized action = "initialized" // 0

// for pre-commit git hook.
var preCommitHook action = "pre-commit" // -1

// for commit-msg git hook.
var commitMsgHook action = "commit-msg" // -2

// for pre-push git hook.
var prePushHook action = "pre-push" // -3

var (
	Clean action = "clean" // 1
	Lint  action = "lint"  // 2
	Test  action = "test"  // 3
	Build action = "build" // 4
)

var (
	instance *Builder
	once     sync.Once
)

type Builder struct {
	fsm     *fsm.FSM
	project *project
	repo    *git.Repository
	hook    action
	report  *Report
	*buildOption
}

func NewBuilder(dir ...string) *Builder {
	once.Do(func() {
		var hook string
		flag.StringVar(&hook, "hook", "", "git hook trigger")
		flag.Parse()
		// corner case: no hook parameter
		if len(hook) > 0 {
			log.Println(color.GreenString("triggered by %s", hook))
		}
		root := moduleDir(dir...)
		repo, err := git.PlainOpen(root)
		if err != nil {
			log.Println(color.YellowString("project is not at version control"))
		}
		instance = &Builder{
			fsm.NewFSM(string(initialized), events(), callBacks()),
			newProject(root),
			repo,
			action(hook),
			&Report{},
			defaultOption(),
		}
	})
	return instance
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

func events() fsm.Events {
	return fsm.Events{
		// builtin for git commit
		{string(preCommitHook), []string{string(initialized)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(initialized)}, string(commitMsgHook)},
		{string(prePushHook), []string{string(initialized)}, string(prePushHook)},
		{string(Clean), []string{string(initialized), string(prePushHook)}, string(Clean)},
		{string(Lint), []string{string(initialized), string(Clean), string(commitMsgHook), string(prePushHook), string(Test)}, string(Lint)},
		{string(Test), []string{string(initialized), string(Clean), string(commitMsgHook), string(prePushHook), string(Lint)}, string(Test)},
		{string(Build), []string{string(Test)}, string(Build)},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// Clean
		string(Clean): func(ctx context.Context, event *fsm.Event) {
			instance.project.clean()
		},
		fmt.Sprintf("after_%s", Clean): func(ctx context.Context, event *fsm.Event) {
			log.Println("clean success")
		},
		// pre-commit: do linter format
		string(preCommitHook): func(ctx context.Context, event *fsm.Event) {
			linter.Scan(instance.project.ModuleDir(), true)
		},
		// commit-msg : validate message
		string(commitMsgHook): func(ctx context.Context, event *fsm.Event) {
			githook.CommitMsg(string(instance.buildOption.MsgPattern))
		},
		// pre-push : validate repo status
		string(prePushHook): func(ctx context.Context, event *fsm.Event) {
			log.Println("prePushHook")
		},
		// Lint
		string(Lint): func(ctx context.Context, event *fsm.Event) {
			if len(instance.hook) > 0 {
				linter.Scan(instance.project.ModuleDir(), true)
			} else {
				linter.Scan(instance.project.ModuleDir(), false)
			}
		},
		fmt.Sprintf("after_%s", string(Lint)): func(ctx context.Context, event *fsm.Event) {
			// @todo validate if there is githook flag
			instance.report.GenLinterReport()
		},
		// Test
		string(Test): func(ctx context.Context, event *fsm.Event) {
			instance.project.test()
		},
		fmt.Sprintf("after_%s", string(Test)): func(ctx context.Context, event *fsm.Event) {
			instance.report.GenTestReport()
		},
		// Build
		string(Build): func(ctx context.Context, event *fsm.Event) {
			instance.project.build()
		},
	}
}

func (builder *Builder) ScriptDir() string {
	return builder.project.ScriptDir()
}

func (builder *Builder) RootDir() string {
	return builder.project.ModuleDir()
}

func (builder *Builder) Run(actions ...action) {
	actions = resort(builder.hook, actions...)
	if len(actions) < 1 {
		log.Println(color.YellowString("no action provided"))
		return
	}
	log.Printf("actions are: %+v \n", actions)
	ctx := context.WithValue(context.Background(), "report", &Report{})
	for _, evt := range actions {
		err := builder.fsm.Event(ctx, string(evt))
		if err != nil {
			log.Fatalln(color.RedString("%v", err))
		}
	}
}

func resort(builtIn action, actions ...action) []action {
	switch builtIn {
	case preCommitHook:
		return []action{preCommitHook}
	case commitMsgHook:
		return []action{commitMsgHook, Clean, Test, Lint}
	case prePushHook:
		return []action{prePushHook, Clean, Test, Lint}
	default:
		var r []action
		for _, a := range actions {
			if a == Build {
				r = append(r, Test, a)
			} else {
				r = append(r, a)
			}
		}
		return r
	}
}
