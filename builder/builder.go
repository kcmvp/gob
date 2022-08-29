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
	"github.com/kcmvp/gob/infra"
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
	SetupGitHook  Action = "SetupGitHook"
	SetupBuilder  Action = "SetupBuilder"
	Clean         Action = "clean"
	Lint          Action = "lint"
	Test          Action = "test"
	Build         Action = "build"
)

var actionResultMap = map[Action][]string{
	Test: {testCoverOut, testPackageCover, testCoverReport},
	Lint: {linterReport, linterOut},
}

func RunAction(v string) (Action, bool) {
	for _, action := range []Action{Clean, Lint, Test, Build} {
		if string(action) == v {
			return action, true
		}
	}
	return "", false
}

var (
	instance *Builder
	once     sync.Once
)

type Builder struct {
	fsm *fsm.FSM
	*buildOption
	root      string
	scriptDir string
	targetDir string
	hook      Action
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
			defaultOption(),
			root,
			filepath.Join(root, projectScriptDir),
			filepath.Join(root, projectTargetDir),
			hook,
		}
		ctx := context.WithValue(context.Background(), infra.ProjectRootDir, root)
		infra.SetupHookService(ctx)
		infra.SetupLinterService(ctx)
	})
	return instance
}

//nolint:govet
func events() fsm.Events {
	return fsm.Events{
		// builtin for git commit
		{string(SetupGitHook), []string{string(initialized)}, string(SetupGitHook)},
		{string(SetupBuilder), []string{string(initialized)}, string(SetupBuilder)},
		{string(preCommitHook), []string{string(Test)}, string(preCommitHook)},
		{string(commitMsgHook), []string{string(SetupGitHook)}, string(commitMsgHook)},
		{string(prePushHook), []string{string(initialized), string(Test)}, string(prePushHook)},
		{string(Clean), []string{string(initialized), string(prePushHook), string(SetupGitHook), string(Lint), string(Test), string(Build)}, string(Clean)},
		// commit-msg and pre-push can't trigger lint
		{string(Lint), []string{string(initialized), string(SetupGitHook), string(Clean), string(preCommitHook), string(Test)}, string(Lint)},
		{string(Test), []string{string(initialized), string(SetupGitHook), string(Clean), string(commitMsgHook), string(prePushHook), string(Lint)}, string(Test)},
		{string(Build), []string{string(Test), string(Clean)}, string(Build)},
	}
}

func callBacks() fsm.Callbacks {
	return map[string]fsm.Callback{
		// setup builder
		fmt.Sprintf("before_%s", SetupBuilder): createDirCallback,
		string(SetupBuilder):                   setupBuilderCallback,
		// setup git
		fmt.Sprintf("before_%s", SetupGitHook): createDirCallback,
		string(SetupGitHook):                   gitHookCallback,
		// Clean
		string(Clean):                  cleanCallback,
		fmt.Sprintf("after_%s", Clean): afterCleanCallback,
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

func (builder *Builder) ScriptDir() string {
	return builder.scriptDir
}

func (builder *Builder) TargetDir() string {
	return builder.targetDir
}

func (builder *Builder) RootDir() string {
	return builder.root
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
			log.Fatalln(color.RedString("%s", err.Error()))
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

func CleanResult(a Action) {
	if _, err := os.Stat(instance.TargetDir()); err != nil {
		return
	}
	if outputs, ok := actionResultMap[a]; ok {
		log.Printf("Cleaning %s output \n", a)
		for _, output := range outputs {
			os.Remove(filepath.Join(instance.TargetDir(), output))
		}
	}
}
