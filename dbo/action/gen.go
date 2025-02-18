package action

import (
	"fmt"
	"github.com/kcmvp/app"
	"github.com/kcmvp/dbo/scaffold/meta"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"path/filepath"
)

func validateConfig(cmd *cobra.Command, args []string) error {
	return lo.If(len(meta.Platforms()) < 1, fmt.Errorf("can not find datasource in application.yml")).Else(nil)
}

func generate(cmd *cobra.Command, args []string) error {
	if diagram := meta.Build(); diagram.IsOk() {
		return diagram.MustGet().ER(filepath.Join(app.RootDir(), "target"))
	} else {
		return diagram.Error()
	}
}

// genCmd Generate ER diagram, schema for project
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate ER diagram and schema for project",
	Long: `Generate ER diagram and schema for project.
An entity struct must implement github.com/kcmvp/dbo/base/IEntity
`,
	PersistentPreRunE: validateConfig,
	RunE:              generate,
}

func init() {
	rootCmd.AddCommand(genCmd)
}
