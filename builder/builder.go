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

	"github.com/fatih/color"
	"github.com/looplab/fsm"
	"golang.org/x/mod/modfile"
)

const (
	projectScriptDir = "scripts"
	projectTargetDir = "target"
	testCoverOut     = "cover.out"
	testCoverReport  = "cover.html"
	testPackageCover = "cover_package.json"
	linterReport     = "golangci-lint.html"
	linterOut        = "golangci-lint.out"
)

type ContextKey string

const BuilderContextKey ContextKey = "_builder"

const (
	ScanAll = "_scan_all"
	GenHook = "_gen_hook"
)

type Action string

var (
	initialized   Action = "initialized"
	preCommitHook Action = "pre-commit"
	commitMsgHook Action = "commit-msg"
	prePushHook   Action = "pre-push"
	GenGitHook    Action = "genGitHook"
	GenBuilder    Action = "genBuilder"
	Clean         Action = "clean"
	Lint          Action = "lint"
	Test          Action = "test"
	Build         Action = "build"
)

func actionOutputMap(action Action) []string {
	m := map[Action][]string{
		Test: {testCoverOut, testPackageCover, testCoverReport},
		Lint: {linterReport, linterOut},
	}
	return m[action]
}

func RunAction(v string) (Action, bool) {
	for _, action := range []Action{Clean, Lint, Test, Build} {
		if string(action) == v {
			return action, true
		}
	}
	return "", false
}

// var (
//	builder *Builder
//)

type Builder struct {
	fsm *fsm.FSM
	*buildOption
	Buildable
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

func NewBuilderWith(project Buildable) *Builder {
	data, err := os.ReadFile(filepath.Join(project.RootDir(), "go.mod"))
	if err != nil {
		log.Fatalln(color.RedString("can not find go.mod in %s", project.RootDir()))
	} else {
		if _, err := modfile.Parse("go.mod", data, nil); err != nil {
			log.Fatalln(color.RedString("invalid mod file %v", err))
		}
	}
	var hook Action
	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(2, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		hook = trigger(filepath.Base(frame.File))
		if len(hook) > 0 {
			log.Printf("%s \n", hook)
			break
		}
	}
	builder := &Builder{
		fsm.NewFSM(string(initialized), builderEvents(), callBacks()),
		defaultOption(),
		project,
		hook,
	}
	return builder
}

func NewBuilder(root string) *Builder {
	return NewBuilderWith(NewDefaultBuildable(root))
}

//nolint:govet
func builderEvents() fsm.Events {
	return fsm.Events{
		// builtin for git commit
		{string(GenGitHook), []string{string(initialized)}, string(GenGitHook)},
		{string(GenBuilder), []string{string(initialized)}, string(GenBuilder)},
		{string(preCommitHook), []string{string(Test)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(GenGitHook)}, string(commitMsgHook)},
		{string(prePushHook), []string{string(initialized), string(Test)}, string(prePushHook)},
		{string(Clean), []string{string(initialized), string(prePushHook), string(GenGitHook), string(Lint), string(Test), string(Build)}, string(Clean)},
		// commit-msg and pre-push can't trigger lint
		{string(Lint), []string{string(initialized), string(GenGitHook), string(Clean), string(preCommitHook), string(Test)}, string(Lint)},
		{string(Test), []string{string(initialized), string(GenGitHook), string(Clean), string(commitMsgHook), string(prePushHook), string(Lint)}, string(Test)},
		{string(Build), []string{string(Test), string(Clean)}, string(Build)},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// setup builder
		fmt.Sprintf("before_%s", GenBuilder): createDirCallback,
		string(GenBuilder):                   genBuilderCallback,
		// setup git
		fmt.Sprintf("before_%s", GenGitHook): createDirCallback,
		string(GenGitHook):                   genGitHookCallback,
		// Clean
		string(Clean):                  cleanAllCallback,
		fmt.Sprintf("after_%s", Clean): afterCleanAllCallback,
		// pre-commit
		string(preCommitHook): preCommitCallback,
		// commit-msg
		string(commitMsgHook): commitMsgCallback,
		// pre-push
		string(prePushHook): prePushHookCallback,
		// lint
		fmt.Sprintf("before_%s", Lint): createDirCallback,
		string(Lint):                   lintCallback,
		// test
		fmt.Sprintf("before_%s", Test): createDirCallback,
		string(Test):                   testCallback,
		// build
		string(Build): buildCallback,
	}
}

func (builder *Builder) Run(actions ...Action) {
	ctx := context.WithValue(context.Background(), BuilderContextKey, builder)
	RunCtx(ctx, actions...)
}

func RunCtx(ctx context.Context, actions ...Action) {
	builder := GetBuilder(ctx)
	actions = sort(builder.hook, actions...)
	if len(actions) < 1 {
		log.Println(color.YellowString("no Action provided"))
	}
	ctx = context.WithValue(ctx, BuilderContextKey, builder)
	for _, evt := range actions {
		err := builder.fsm.Event(ctx, string(evt))
		var t1 fsm.CanceledError
		if err != nil && !errors.Is(err, t1) {
			log.Fatalln(color.RedString("%s", err.Error()))
		}
	}
}

func sort(builtIn Action, actions ...Action) []Action {
	switch builtIn {
	case preCommitHook:
		return []Action{GenGitHook, Lint, Test, preCommitHook}
	case commitMsgHook:
		return []Action{GenGitHook, commitMsgHook}
	case prePushHook:
		return []Action{GenGitHook, Test, prePushHook}
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

func (builder *Builder) cleanOutput(a Action) {
	if _, err := os.Stat(builder.TargetDir()); err != nil {
		return
	}
	log.Printf("Cleaning %s output \n", a)
	for _, output := range actionOutputMap(a) {
		os.Remove(filepath.Join(builder.TargetDir(), output))
	}
}

func GetBuilder(ctx context.Context) *Builder {
	b, ok := ctx.Value(BuilderContextKey).(*Builder)
	if ok {
		return b
	}
	log.Fatalln("Failed to get builder from context")
	return nil
}

func cancelEvent(event *fsm.Event, err error, msg ...string) {
	if err != nil {
		event.Cancel(fmt.Errorf("%s:%w", msg, err))
	}
}
