package builder

import (
	"context"
	"fmt"
	"github.com/kcmvp/gos/infra"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/looplab/fsm"
	"golang.org/x/mod/modfile"
)

const (
	ScanAll = "_scan_all"
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
		repo, err := git.PlainOpen(root)
		if err != nil {
			log.Println(color.YellowString("project is not at version control"))
		}
		var hook Action
		_, filename, _, ok := runtime.Caller(4)
		if ok {
			hook = trigger(filepath.Base(filename))
			if len(hook) > 0 {
				log.Println(color.GreenString("%s", hook))
			}
		}
		instance = &Builder{
			fsm.NewFSM(string(initialized), events(), callBacks()),
			newProject(root),
			repo,
			defaultOption(),
			root,
			hook,
		}
		instance.setup()
	})
	return instance
}

func events() fsm.Events {
	return fsm.Events{
		// builtin for git commit
		{string(preCommitHook), []string{string(initialized)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(initialized)}, string(commitMsgHook)},
		{string(prePushHook), []string{string(initialized)}, string(prePushHook)},
		{string(Clean), []string{string(initialized), string(prePushHook)}, string(Clean)},
		// commit-msg and pre-push can't trigger lint
		{string(Lint), []string{string(initialized), string(Clean), string(preCommitHook), string(Test)}, string(Lint)},
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
			//linter.LintScan(instance.project.TargetDir(), true)
			log.Println("run linter against project")
		},
		// commit-msg : validate message
		string(commitMsgHook): func(ctx context.Context, event *fsm.Event) {
			infra.CommitMsg(string(instance.buildOption.MsgPattern))
		},
		// pre-push : validate repo status
		string(prePushHook): func(ctx context.Context, event *fsm.Event) {
			log.Println("validate unit test coverage")
			//if instance.repo != nil {
			//	infra.PrePush(filepath.Join(instance.root, scriptDir, "coverage.json"),
			//		filepath.Join(instance.root, targetDir, "coverage.json"), instance.repo)
			//}
		},
		// Lint
		string(Lint): func(ctx context.Context, event *fsm.Event) {
			v, ok := ctx.Value(ScanAll).(bool)
			if !ok {
				v = false
			}
			infra.LintScan(instance.project.TargetDir(), v)
		},
		fmt.Sprintf("after_%s", string(Lint)): func(ctx context.Context, event *fsm.Event) {
			infra.VerifyLinter(string(preCommitHook) == event.Src)
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
		return
	}
	log.Printf("actions are: %+v \n", actions)
	for _, evt := range actions {
		err := builder.fsm.Event(ctx, string(evt))
		if err != nil {
			log.Fatalln(color.RedString("%v", err))
		}
	}
}

func sort(builtIn Action, actions ...Action) []Action {
	switch builtIn {
	case preCommitHook:
		return []Action{preCommitHook, Lint, Test}
	case commitMsgHook:
		return []Action{commitMsgHook}
	case prePushHook:
		return []Action{Test, prePushHook}
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

func (builder *Builder) setup() {
	if builder.repo != nil {
		if infra.SetupHook(builder.RootDir(), builder.ScriptDir(), false) != nil {
			log.Println(color.RedString("failed to setup hook"))
		}
	}
}
