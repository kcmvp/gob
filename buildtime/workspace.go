package buildtime

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dominikbraun/graph"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/common"
	"github.com/samber/lo"
	_ "github.com/samber/lo/parallel"
	"github.com/samber/mo"
	"go/types"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ws workspace[string, *Module]
)

type workspace[K comparable, T any] struct {
	graph.Graph[K, T]
	dir2Path map[string]string
}

type Module struct {
	dir          string
	mod          *modfile.File
	ver          string
	latestVer    string
	dependencies []Module
}

func (module *Module) Path() string {
	return module.mod.Module.Mod.Path
}

func (module *Module) Dir() string {
	return module.dir
}

func (module *Module) MainFile() mo.Option[string] {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedFiles | packages.NeedSyntax,
		Dir:  module.Dir(),
	}
	pkgs := mo.TupleToResult(packages.Load(cfg, "./...")).MustGet()
	files := lo.FilterMap(pkgs, func(pkg *packages.Package, _ int) (string, bool) {
		if pkg.Name != "main" {
			return "", false
		}
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if f, ok := obj.(*types.Func); ok {
				signature := f.Type().(*types.Signature)
				if f.Name() == "main" && signature.Params().Len() == 0 && signature.Results().Len() == 0 {
					return pkg.Fset.Position(obj.Pos()).Filename, true
				}
			}
		}
		return "", false
	})
	return lo.If(len(files) == 0, mo.None[string]()).ElseF(func() mo.Option[string] {
		return mo.Some(files[0])
	})
}

// init initialize ws
func init() {
	ws = workspace[string, *Module]{
		Graph: graph.New[string, *Module](func(module *Module) string {
			return module.Path()
		}, graph.Directed(), graph.PreventCycles()),
		dir2Path: map[string]string{},
	}
	rs := mo.TupleToResult(exec.Command("go", "list", "-m", "-f", "{{.Dir}}").CombinedOutput())
	if rs.IsError() || len(rs.MustGet()) == 0 {
		log.Fatal(color.RedString("please execute command in ws or project root directory"))
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(rs.MustGet()))
	for scanner.Scan() {
		dir := strings.TrimSpace(scanner.Text())
		data := mo.TupleToResult(os.ReadFile(filepath.Join(dir, "go.mod"))).MustGet()
		mod := mo.TupleToResult(modfile.Parse("go.mod", data, nil)).MustGet()
		ws.AddVertex(&Module{dir: dir, mod: mod})
		ws.dir2Path[dir] = mod.Module.Mod.Path
	}
	vertices, _ := ws.PredecessorMap()
	if len(vertices) > 1 {
		lo.ForEach(lo.Keys(vertices), func(path string, index int) {
			// build reference
			module := mo.TupleToResult(ws.Vertex(path)).MustGet()
			lo.ForEach(module.mod.Require, func(item *modfile.Require, index int) {
				if mo.TupleToResult(ws.Vertex(item.Mod.Path)).IsOk() {
					ws.AddEdge(path, item.Mod.Path, graph.EdgeAttribute("ref", "1"))
				}
			})
		})
	}
}

func CurrentModule() mo.Result[*Module] {
	if path, ok := ws.dir2Path[common.CurrentDir()]; ok {
		return mo.Ok(mo.TupleToResult(ws.Vertex(path)).MustGet())
	}
	return mo.Errf[*Module]("please execute the command in the ws or project root")
}

func Modules() []*Module {
	return lo.Map(lo.Values(ws.dir2Path), func(path string, index int) *Module {
		m, _ := ws.Vertex(path)
		return m
	})
}

func RootDir() string {
	paths := lo.Keys(ws.dir2Path)
	rootSections := strings.FieldsFunc(paths[0], func(r rune) bool { return r == os.PathSeparator })
	var root string
	lo.ForEachWhile(rootSections, func(section string, idx int) bool {
		tmp := strings.Join(rootSections[:idx+1], string(os.PathSeparator))
		if !common.WindowsEnv() {
			tmp = fmt.Sprintf("%s%s", string(os.PathSeparator), tmp)
		}
		if lo.EveryBy(paths, func(path string) bool {
			return strings.HasPrefix(path, tmp)
		}) {
			root = tmp
			return true
		}
		return false
	})
	return root
}
