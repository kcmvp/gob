package command

import (
	"fmt"
	"golang.org/x/mod/modfile" //nolint
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

func directLatest() []lo.Tuple2[string, string] {
	exec.Command("go", "mod", "tidy").CombinedOutput() //nolint
	directs := lo.FilterMap(artifact.CurProject().Dependencies(), func(item *modfile.Require, _ int) (lo.Tuple2[string, string], bool) {
		return lo.Tuple2[string, string]{A: item.Mod.Path, B: item.Mod.Version}, !item.Indirect
	})
	return artifact.LatestVersion(false, lo.Map(directs, func(item lo.Tuple2[string, string], _ int) string {
		return item.A
	})...)
}

func upgradeAll() error {
	candidates := lo.Filter(directLatest(), func(latest lo.Tuple2[string, string], _ int) bool {
		return lo.ContainsBy(artifact.CurProject().Dependencies(), func(dependency *modfile.Require) bool {
			return !dependency.Indirect && dependency.Mod.Path == latest.A && dependency.Mod.Version != latest.B
		})
	})
	args := lo.Union([]string{"get", "-u"}, lo.Map(candidates, func(latest lo.Tuple2[string, string], _ int) string {
		return latest.A
	}))
	cmd := exec.Command("go", args...)
	if err := artifact.PtyCmdOutput(cmd, "upgrading dependencies ......", false, nil); err != nil {
		color.Red("failed to upgrade dependencies: %s", err.Error())
	}
	exec.Command("go", "mod", "tidy").CombinedOutput() //nolint
	return nil
}

// dependencyTree build dependency tree of the project, an empty tree returns when runs into error
func dependencyTree() (treeprint.Tree, error) {
	tree := treeprint.New()
	tree.SetValue(artifact.CurProject().Module())
	// get the latest version
	versions := directLatest()
	// parse the dependency tree
	cache := []string{os.Getenv("GOPATH"), "pkg", "mod", "cache", "download"}
	for _, dependency := range artifact.CurProject().Dependencies() {
		if !dependency.Indirect {
			m, ok := lo.Find(versions, func(item lo.Tuple2[string, string]) bool {
				return dependency.Mod.Path == item.A && dependency.Mod.Version != item.B
			})
			label := lo.IfF(!ok, func() string {
				return dependency.Mod.String()
			}).ElseF(func() string {
				return yellow.Sprintf("* %s (%s)", dependency.Mod.String(), m.B)
			})
			direct := tree.AddBranch(label)
			dir := append(cache, strings.Split(dependency.Mod.Path, "/")...)
			dir = append(dir, []string{"@v", fmt.Sprintf("%s.mod", dependency.Mod.Version)}...)
			data, err := os.ReadFile(filepath.Join(dir...))
			if err != nil {
				color.Yellow("failed to get latest version of %s", dependency.Mod.Path)
				continue
			}
			mod, _ := modfile.Parse("go.mod", data, nil)
			lo.ForEach(artifact.CurProject().Dependencies(), func(c *modfile.Require, index int) {
				if c.Indirect && lo.ContainsBy(mod.Require, func(m *modfile.Require) bool {
					return !m.Indirect && c.Mod.Path == m.Mod.Path
				}) {
					direct.AddNode(c.Mod.String())
				}
			})
		}
	}
	return tree, nil
}

// depCmd represents the dep command
var depCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show the dependency tree of the project",
	Long: `Show the dependency tree of the project
and indicate available updates which take an green * indicator`,
	RunE: func(cmd *cobra.Command, args []string) error {
		upgrade, _ := cmd.Flags().GetBool("upgrade")
		if upgrade {
			return upgradeAll()
		}
		tree, err := dependencyTree()
		if err != nil {
			return err
		} else if tree == nil {
			yellow.Println("No dependencies !")
			return nil
		}
		fmt.Println("\rDependencies of the projects:")
		fmt.Println(tree.String())
		return nil
	},
}

func init() {
	depCmd.SetUsageTemplate(usageTemplate())
	depCmd.SetErrPrefix(color.RedString("Error:"))
	depCmd.Flags().BoolP("upgrade", "u", false, "upgrade dependencies if outdated dependencies exist")
	rootCmd.AddCommand(depCmd)
}
