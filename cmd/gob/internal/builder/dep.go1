package builder

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/gob/cmd/gob/project"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

func dependencyTree() (treeprint.Tree, error) {
	tree := treeprint.New()
	tree.SetValue(fmt.Sprintf("Dependenies of %s", project.Module()))
	for _, dep := range project.Dependencies() {
		label := lo.IfF(dep.Ver() != dep.LatestVer(), func() string {
			return color.YellowString("%s(%s)", dep.String(), dep.LatestVer())
		}).Else(dep.String())
		branch := tree.AddBranch(label)
		lo.ForEach(dep.Dependencies(), func(c project.Dependency, _ int) {
			branch.AddNode(c.String())
		})
	}
	return tree, nil
}

func DepCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dep",
		Short: color.GreenString(`List or update project's dependencies'`),
		Long:  color.GreenString(`List or update project's dependencies'`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if tv, err := dependencyTree(); err != nil {
				return err
			} else {
				fmt.Println(tv.String())
			}
			return nil
		},
	}
}
