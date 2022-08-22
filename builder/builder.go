package builder

import (
	"context"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/kcmvp/gos/builder/githook"
	"github.com/kcmvp/gos/builder/linter"
	"github.com/looplab/fsm"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Action string

var initialized Action = "initialized" // 0

// for pre-commit git hook.
var preCommitHook Action = "pre-commit" // -1

// for commit-msg git hook.
var commitMsgHook Action = "commit-msg" // -2

// for pre-push git hook.
var prePushHook Action = "pre-push" // -3

var (
	Clean Action = "clean" // 1
	Lint  Action = "lint"  // 2
	Test  Action = "test"  // 3
	Build Action = "build" // 4
)

func ValueOf(v string) (Action, bool) {
	am := map[string]Action{
		"clean": Clean,
		"lint":  Lint,
		"test":  Test,
		"build": Build,
	}
	a, ok := am[v]
	return a, ok
}

var (
	instance *Builder
	once     sync.Once
)

type Builder struct {
	fsm     *fsm.FSM
	project *project
	repo    *git.Repository
	hook    Action
	report  *Report
	*buildOption
}

func NewBuilder(root string) *Builder {
	once.Do(func() {
		data, err := os.ReadFile(filepath.Join(root, "go.mod"))
		if err != nil {
			log.Fatalln(color.RedString("can not find go.mod in %s", root))
		} else {
			if _, err := modfile.Parse("go.mod", data, nil); err != nil {
				log.Fatalln(color.RedString("invalid mod file %v", err))
			}
		}

		var hook string
		flag.StringVar(&hook, "hook", "", "git hook trigger")
		flag.Parse()
		// corner case: no hook parameter
		if len(hook) > 0 {
			log.Println(color.GreenString("triggered by %s", hook))
		}
		repo, err := git.PlainOpen(root)
		if err != nil {
			log.Println(color.YellowString("project is not at version control"))
		}
		instance = &Builder{
			fsm.NewFSM(string(initialized), events(), callBacks()),
			newProject(root),
			repo,
			Action(hook),
			&Report{},
			defaultOption(),
		}
	})
	return instance
}

//func moduleDir(dir ...string) string {
//	if len(dir) > 0 {
//		if _, err := os.ReadFile(filepath.Join(dir[0], "go.mod")); err == nil {
//			return dir[0]
//		}
//	}
//	_, file, _, ok := runtime.Caller(2)
//	if ok {
//		p := filepath.Dir(file)
//		for p != string(os.PathSeparator) {
//			if _, err := os.ReadFile(filepath.Join(p, "go.mod")); err == nil {
//				return p
//			} else {
//				p = filepath.Dir(p)
//			}
//		}
//	}
//	panic(color.RedString("can't figure out project mod file"))
//}

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
			linter.Scan(instance.project.ModuleDir(), true, true)
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
				linter.Scan(instance.project.TargetDir(), true, false)
			} else {
				linter.Scan(instance.project.TargetDir(), false, false)
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

func (builder *Builder) Run(actions ...Action) {
	actions = resort(builder.hook, actions...)
	if len(actions) < 1 {
		log.Println(color.YellowString("no Action provided"))
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

func resort(builtIn Action, actions ...Action) []Action {
	switch builtIn {
	case preCommitHook:
		return []Action{preCommitHook}
	case commitMsgHook:
		return []Action{commitMsgHook, Clean, Test, Lint}
	case prePushHook:
		return []Action{prePushHook, Clean, Test, Lint}
	default:
		var r []Action
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
