package command

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kcmvp/gob/cmd/gbc/artifact" //nolint

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

var (
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
)

// parseMod return a tuple which the fourth element is the indicator of direct or indirect reference
func parseMod(mod *os.File) (string, string, []*lo.Tuple4[string, string, string, int], error) {
	scanner := bufio.NewScanner(mod)
	var deps []*lo.Tuple4[string, string, string, int]
	var module, version string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line == ")" || line == "//" || strings.HasPrefix(line, "require") {
			continue
		}
		if strings.HasPrefix(line, "module ") {
			module = strings.Split(line, " ")[1]
		} else if strings.HasPrefix(line, "go ") {
			version = strings.Split(line, " ")[1]
		} else {
			entry := strings.Split(line, " ")
			m := strings.TrimSpace(entry[0])
			v := strings.TrimSpace(entry[1])
			dep := lo.T4(m, v, v, lo.If(len(entry) > 2, 0).Else(1))
			deps = append(deps, &dep)
		}
	}
	return module, version, deps, scanner.Err()
}

// dependencyTree build dependency tree of the project, an empty tree returns when runs into error
func dependencyTree() (treeprint.Tree, error) {
	mod, err := os.Open(filepath.Join(artifact.CurProject().Root(), "go.mod"))
	if err != nil {
		return nil, fmt.Errorf(color.RedString(err.Error()))
	}
	exec.Command("go", "mod", "tidy").CombinedOutput() //nolint
	if output, err := exec.Command("go", "build", "./...").CombinedOutput(); err != nil {
		return nil, fmt.Errorf(color.RedString(string(output)))
	}
	module, _, dependencies, err := parseMod(mod)
	if len(dependencies) < 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	tree := treeprint.New()
	tree.SetValue(module)
	direct := lo.FilterMap(dependencies, func(item *lo.Tuple4[string, string, string, int], _ int) (string, bool) {
		return item.A, item.D == 1
	})
	// get the latest version
	versions := artifact.LatestVersion(direct...)
	for _, dep := range dependencies {
		if version, ok := lo.Find(versions, func(t lo.Tuple2[string, string]) bool {
			return dep.A == t.A && dep.B != t.B
		}); ok {
			dep.C = version.B
		}
	}
	// parse the dependency tree
	cache := []string{os.Getenv("GOPATH"), "pkg", "mod", "cache", "download"}
	for _, dependency := range dependencies {
		if dependency.D == 1 {
			label := lo.IfF(dependency.B == dependency.C, func() string {
				return fmt.Sprintf("%s@%s", dependency.A, dependency.B)
			}).ElseF(func() string {
				return yellow.Sprintf("* %s@%s (%s)", dependency.A, dependency.B, dependency.C)
			})
			child := tree.AddBranch(label)
			dir := append(cache, strings.Split(dependency.A, "/")...)
			dir = append(dir, []string{"@v", fmt.Sprintf("%s.mod", dependency.B)}...)
			mod, err = os.Open(filepath.Join(dir...))
			if err != nil {
				return tree, fmt.Errorf(color.RedString(err.Error()))
			}
			_, _, cDeps, err := parseMod(mod)
			if err != nil {
				return tree, fmt.Errorf(color.RedString(err.Error()))
			}
			inter := lo.Filter(cDeps, func(c *lo.Tuple4[string, string, string, int], _ int) bool {
				return lo.ContainsBy(dependencies, func(p *lo.Tuple4[string, string, string, int]) bool {
					return p.A == c.A
				})
			})
			for _, l := range inter {
				child.AddNode(fmt.Sprintf("%s@%s", l.A, l.B))
			}
		}
	}
	return tree, err
}

// depCmd represents the dep command
var depCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show the dependency tree of the project",
	Long: `Show the dependency tree of the project
and indicate available updates which take an green * indicator`,
	RunE: func(cmd *cobra.Command, args []string) error {
		tree, err := dependencyTree()
		if err != nil {
			return err
		} else if tree == nil {
			yellow.Println("No dependencies !")
			return nil
		}
		green.Println("Dependencies of the projects:")
		fmt.Println(tree.String())
		return nil
	},
}

func init() {
	depCmd.SetUsageTemplate(usageTemplate())
	depCmd.SetErrPrefix(color.RedString("Error:"))
	rootCmd.AddCommand(depCmd)
}
