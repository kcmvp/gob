package project

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"golang.org/x/mod/modfile"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Dependency struct {
	module       string
	ver          string
	latestVer    string
	dependencies []Dependency
}

func (dep Dependency) Module() string {
	return dep.module
}

func (dep Dependency) Ver() string {
	return dep.ver
}

func (dep Dependency) LatestVer() string {
	return dep.latestVer
}

func (dep Dependency) String() string {
	return fmt.Sprintf("%s@%s", dep.module, dep.ver)
}
func (dep Dependency) Dependencies() []Dependency {
	return dep.dependencies
}

func Dependencies() []Dependency {
	exec.Command("go", "mod", "tidy").CombinedOutput() //nolint
	cache := []string{os.Getenv("GOPATH"), "pkg", "mod", "cache", "download"}
	return lo.FilterMap(project.mod.Require, func(mod *modfile.Require, index int) (Dependency, bool) {
		if !mod.Indirect {
			dep := Dependency{module: mod.Mod.Path, ver: mod.Mod.Version}
			rs := mo.TupleToResult(exec.Command("go", append([]string{"list", "-m"}, dep.String(), fmt.Sprintf("%s@latest", dep.module))...).CombinedOutput())
			if rs.IsError() {
				color.Yellow("failed to get latest version of %s", mod.Mod.Path)
				return Dependency{}, false
			}
			scanner := bufio.NewScanner(bytes.NewBuffer(rs.MustGet()))
			var latest string
			for scanner.Scan() {
				latest = scanner.Text()
			}
			if entry := strings.Fields(latest); len(entry) > 1 {
				dep.latestVer = entry[1]
			}
			dir := append(cache, strings.Split(mod.Mod.Path, "/")...)
			dir = append(dir, []string{"@v", fmt.Sprintf("%s.mod", mod.Mod.Version)}...)
			if data := mo.TupleToResult(os.ReadFile(filepath.Join(dir...))); data.IsOk() {
				if child := mo.TupleToResult(modfile.Parse("go.mod", data.MustGet(), nil)); child.IsOk() {
					lo.ForEach(project.mod.Require, func(r1 *modfile.Require, index int) {
						if lo.ContainsBy(child.MustGet().Require, func(r2 *modfile.Require) bool {
							return r1.Mod.Path == r2.Mod.Path
						}) {
							dep.dependencies = append(dep.dependencies, Dependency{module: r1.Mod.Path, ver: r1.Mod.Version})
						}
					})
				} else {
					color.Yellow("failed to parse dependency %s", mod.Mod.Path)
				}
			} else {
				color.Yellow("failed to get dependency %s", mod.Mod.Path)
			}
			return dep, true
		}
		return Dependency{}, false
	})
}
