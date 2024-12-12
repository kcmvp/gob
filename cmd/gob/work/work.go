package work

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/core/env"
	"github.com/samber/lo"
	_ "github.com/samber/lo/parallel"
	"github.com/samber/mo"
	"golang.org/x/mod/modfile"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	root    string
	modules graph.Graph[string, *Module]
	pairs   []Pair
)

type Pair lo.Tuple2[string, string]

func (p Pair) Dir() string {
	return p.A
}

func (p Pair) Path() string {
	return p.B
}

type Module struct {
	dir          string
	mainFile     string
	mod          *modfile.File
	ver          string
	latestVer    string
	dependencies []Module
}

func (module *Module) Path() string {
	return lo.If(module.mod == nil, "root").Else(module.mod.Module.Mod.Path)
}

func (module *Module) MultipleModule() bool {
	return module.mod == nil
}

func (module *Module) Root() string {
	return module.dir
}

func Root() string {
	return root
}

// init initialize workspace
func init() {
	modules = graph.New[string, *Module](func(module *Module) string {
		return module.dir
	}, graph.Directed(), graph.PreventCycles())
	rs := mo.TupleToResult(exec.Command("go", "list", "-m", "-f", "{{.Dir}}").CombinedOutput())
	if rs.IsError() || len(rs.MustGet()) == 0 {
		log.Fatal(color.RedString("please execute command in workspace or project root directory"))
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(rs.MustGet()))
	var rootSections []string
	for scanner.Scan() {
		dir := strings.TrimSpace(scanner.Text())
		data := mo.TupleToResult(os.ReadFile(filepath.Join(dir, "go.mod"))).MustGet()
		mod := mo.TupleToResult(modfile.Parse("go.mod", data, nil)).MustGet()
		modules.AddVertex(&Module{dir: dir, mod: mod})
		pairs = append(pairs, Pair{A: dir, B: mod.Module.Mod.Path})
		sections := strings.FieldsFunc(dir, func(r rune) bool { return r == os.PathSeparator })
		rootSections = lo.Intersect(rootSections, sections)
		if len(rootSections) == 0 {
			rootSections = sections
		}
	}
	root = strings.Join(rootSections, string(os.PathSeparator))
	if !env.WindowsEnv() {
		root = fmt.Sprintf("%s%s", string(os.PathSeparator), root)
	}
	if len(pairs) > 1 {
		modules.AddVertex(&Module{dir: root})
		lo.ForEach(pairs, func(pair Pair, _ int) {
			modules.AddEdge(root, pair.Dir())
		})
	}
	// build reference
	lo.ForEach(pairs, func(pp Pair, _ int) {
		module := mo.TupleToResult(modules.Vertex(pp.Dir())).MustGet()
		refs := lo.FilterMap(module.mod.Require, func(item *modfile.Require, _ int) (Pair, bool) {
			return lo.Find(pairs, func(p Pair) bool {
				return p.Path() == item.Mod.Path
			})
		})
		lo.ForEach(refs, func(cp Pair, _ int) {
			if err := modules.AddEdge(pp.Dir(), cp.Dir()); err != nil {
				panic(err.Error())
			}
		})
	})
}

func Current() mo.Result[*Module] {
	rs := mo.TupleToResult(modules.Vertex(mo.TupleToResult(os.Getwd()).MustGet()))
	return lo.If(rs.IsError(),
		mo.Errf[*Module]("please execute the command in the workspace or project root")).
		ElseF(func() mo.Result[*Module] {
			return mo.Ok(rs.MustGet())
		})
}
