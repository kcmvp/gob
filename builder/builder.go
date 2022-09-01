package builder

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fatih/color"
	"golang.org/x/mod/modfile"
)

const (
	projectScriptDir = "scripts"
	projectTargetDir = "target"
	testCoverOut     = "cover.out"
	testCoverReport  = "cover.html"
	testPackageCover = "cover.json"
	linterReport     = "golangci-lint.html"
	linterOut        = "golangci-lint.out"
)

type ContextKey string

const CtxKeyBuilder ContextKey = "_builder"

const (
	ScanAll = "_scan_all"
)

type Builder struct {
	*buildOption
	Buildable
	hook string
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

	builder := &Builder{
		defaultOption(),
		project,
		"",
	}

	commands := commandMap()
	pc := make([]uintptr, 15)   //nolint
	n := runtime.Callers(2, pc) //nolint
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		hook := filepath.Base(frame.File)
		if _, ok := commands[hook]; ok {
			builder.hook = hook
			break
		}
	}

	return builder
}

func NewBuilder(root string) *Builder {
	return NewBuilderWith(NewDefaultBuildable(root))
}

func (builder *Builder) Run(cmds ...string) {
	ctx := context.WithValue(context.Background(), CtxKeyBuilder, builder)
	RunCtx(ctx, cmds...)
}

func RunCtx(ctx context.Context, cmds ...string) {
	builder := GetBuilder(ctx)
	if len(builder.hook) > 0 {
		cmds = []string{builder.hook}
	}
	if len(cmds) < 1 {
		log.Println(color.YellowString("no Action provided"))
	}
	ctx = context.WithValue(ctx, CtxKeyBuilder, builder)
	processCommands(ctx, cmds...)
}

func GetBuilder(ctx context.Context) *Builder {
	b, ok := ctx.Value(CtxKeyBuilder).(*Builder)
	if ok {
		return b
	}
	log.Fatalln(color.RedString("failed to get builder from context"))
	return nil
}
