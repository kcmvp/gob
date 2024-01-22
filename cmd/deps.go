/*
Copyright © 2024 kcmvp <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

var (
	bold   = color.New(color.Bold)
	yellow = color.New(color.FgYellow, color.Bold)
)

// parseMod return a tuple which the fourth element is the indicator of direct or indirect reference
func parseMod(mod *os.File) (string, string, []*lo.Tuple4[string, string, string, int], error) {
	scanner := bufio.NewScanner(mod)
	start := false
	var deps []*lo.Tuple4[string, string, string, int]
	var module, version string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			module = strings.Split(line, " ")[1]
		} else if strings.HasPrefix(line, "go ") {
			version = strings.Split(line, " ")[1]
		}
		start = start && line != ")"
		if start && len(line) > 0 {
			entry := strings.Split(line, " ")
			m := strings.TrimSpace(entry[0])
			v := strings.TrimSpace(entry[1])
			dep := lo.T4(m, v, v, lo.If(len(entry) > 2, 0).Else(1))
			deps = append(deps, &dep)
		}
		start = start || strings.HasPrefix(line, "require")
	}
	return module, version, deps, scanner.Err()
}

// dependency build dependency tree of the project, an empty tree returns when runs into error
func dependency() (treeprint.Tree, error) {
	tree := treeprint.New()
	mod, err := os.Open(filepath.Join(internal.CurProject().Root(), "go.mod"))
	if err != nil {
		return tree, fmt.Errorf(color.RedString(err.Error()))
	}
	if _, err = exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		return tree, fmt.Errorf(color.RedString(err.Error()))
	}

	if _, err = exec.Command("go", "build", "./...").CombinedOutput(); err != nil {
		return tree, fmt.Errorf(color.RedString(err.Error()))
	}

	module, _, dependencies, err := parseMod(mod)
	if err != nil {
		return tree, fmt.Errorf(err.Error())
	}
	tree.SetValue(bold.Sprintf("%s", module))
	direct := lo.FilterMap(dependencies, func(item *lo.Tuple4[string, string, string, int], _ int) (string, bool) {
		return fmt.Sprintf("%s@latest", item.A), item.D == 1
	})
	// get the latest version
	output, _ := exec.Command("go", append([]string{"list", "-m"}, direct...)...).CombinedOutput() //nolint
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		entry := strings.Split(line, " ")
		for _, dep := range dependencies {
			if dep.A == entry[0] && dep.B != entry[1] {
				dep.C = entry[1]
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return tree, fmt.Errorf(err.Error())
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
		tree, err := dependency()
		if err != nil {
			return err
		}
		bold.Print("\nDependencies of the projects:\n")
		yellow.Print("* indicates new versions available\n")
		fmt.Println("")
		fmt.Println(tree.String())
		return nil
	},
}

func init() {
	builderCmd.AddCommand(depCmd)
}
