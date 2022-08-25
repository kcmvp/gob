package builder

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/kcmvp/gos/infra"
	"github.com/looplab/fsm"
	"golang.org/x/mod/modfile"
)

const (
	ScanAll = "_scan_all"
	GenHook = "_gen_hook"
)

type Action string

var initialized Action = "initialized" // 0

// SetupGitHook setup git hook action.
var SetupGitHook Action = "SetupGitHook" // 0

// for pre-commit git hook.
var preCommitHook Action = "pre-commit" // -1

// for commit-msg git hook.
var commitMsgHook Action = "commit-msg" // -2

// for pre-push git hook.
var prePushHook Action = "pre-push" // -3

var (
	Clean Action = "clean" // 2
	Lint  Action = "lint"  // 3
	Test  Action = "test"  // 4
	Build Action = "build" // 5
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
	*buildOption

	root string
	hook Action
}

func trigger(name string) Action {
	for _, action := range []Action{preCommitHook, commitMsgHook, prePushHook} {
		if fmt.Sprintf("%s.go", action) == strings.ReplaceAll(name, "_", "-") {
			return action
		}
	}
	return ""
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
		// repo, err := git.PlainOpen(root)
		// if err != nil {
		//	log.Println(color.YellowString("project is not at version control"))
		//}
		var hook Action
		_, filename, _, ok := runtime.Caller(4)
		if ok {
			hook = trigger(filepath.Base(filename))
			if len(hook) > 0 {
				log.Printf("%s \n", hook)
			}
		}
		instance = &Builder{
			fsm.NewFSM(string(initialized), events(), callBacks()),
			newProject(root),
			defaultOption(),
			root,
			hook,
		}
		infra.NewGitHookService(root)
	})
	return instance
}

//nolint:govet
func events() fsm.Events {
	return fsm.Events{
		// builtin for git commit
		{string(SetupGitHook), []string{string(initialized)}, string(SetupGitHook)},
		{string(preCommitHook), []string{string(Test)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(SetupGitHook)}, string(commitMsgHook)},
		{string(prePushHook), []string{string(initialized)}, string(prePushHook)},
		{string(Clean), []string{string(initialized), string(prePushHook)}, string(Clean)},
		// commit-msg and pre-push can't trigger lint
		{string(Lint), []string{string(initialized), string(SetupGitHook), string(Clean), string(preCommitHook), string(Test)}, string(Lint)},
		{string(Test), []string{string(initialized), string(SetupGitHook), string(Clean), string(commitMsgHook), string(prePushHook), string(Lint)}, string(Test)},
		{string(Build), []string{string(Test)}, string(Build)},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// setupGit
		string(SetupGitHook): func(ctx context.Context, event *fsm.Event) {
			v, ok := ctx.Value(GenHook).(bool)
			if err := infra.SetupHook(ok && v); err != nil {
				log.Fatalln(color.RedString("failed to setup hook: %s", err.Error()))
			}
		},
		// Clean
		string(Clean): func(ctx context.Context, event *fsm.Event) {
			instance.project.clean()
		},
		fmt.Sprintf("after_%s", Clean): func(ctx context.Context, event *fsm.Event) {
			log.Println("clean success")
		},
		// pre-commit: do linter format
		string(preCommitHook): func(ctx context.Context, event *fsm.Event) {
			infra.GitAdd("golangci-lint.json", "coverage.json")
		},
		// commit-msg : validate message
		string(commitMsgHook): func(ctx context.Context, event *fsm.Event) {
			infra.CommitMsg(string(instance.buildOption.MsgPattern))
		},
		// pre-push : validate repo status
		string(prePushHook): func(ctx context.Context, event *fsm.Event) {
			log.Println("validate code quality for push")

			//	filepath.Join(instance.root, targetDir, "coverage.json"), instance.repo)
		},
		// Lint : before
		string(Lint): func(ctx context.Context, event *fsm.Event) {
			isPreCommitHooK := preCommitHook == instance.hook
			v, _ := ctx.Value(ScanAll).(bool)
			v = v && !isPreCommitHooK
			infra.LintScan(instance.project.TargetDir(), v, isPreCommitHooK)
			if !isPreCommitHooK {
				event.Cancel()
			}
		},
		// Lint: after
		fmt.Sprintf("after_%s", string(Lint)): func(ctx context.Context, event *fsm.Event) {
			infra.LintScan(instance.project.moduleDir, true, false)
		},
		// Test
		string(Test): func(ctx context.Context, event *fsm.Event) {
			instance.project.test()
		},
		fmt.Sprintf("after_%s", string(Test)): func(ctx context.Context, event *fsm.Event) {
			instance.project.coverage(preCommitHook == instance.hook)
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
	builder.RunCtx(context.Background(), actions...)
}

func (builder *Builder) RunCtx(ctx context.Context, actions ...Action) {
	actions = sort(builder.hook, actions...)
	if len(actions) < 1 {
		log.Println(color.YellowString("no Action provided"))
	}

	for _, evt := range actions {
		err := builder.fsm.Event(ctx, string(evt))
		var t1 fsm.CanceledError
		if err != nil && !errors.Is(err, t1) {
			log.Fatalln(color.RedString("%v", err))
		}
	}
}

func sort(builtIn Action, actions ...Action) []Action {
	switch builtIn {
	case preCommitHook:
		return []Action{SetupGitHook, Lint, Test, preCommitHook}
	case commitMsgHook:
		return []Action{SetupGitHook, commitMsgHook}
	case prePushHook:
		return []Action{SetupGitHook, Test, prePushHook}
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
