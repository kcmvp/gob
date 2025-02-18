package action

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/kcmvp/app"
	"github.com/kcmvp/dbo/scaffold/meta"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func initDB(cmd *cobra.Command, args []string) error {
	result := mo.TupleToResult(modfile.Parse("go.mod", mo.TupleToResult(os.ReadFile(filepath.Join(app.RootDir(), "go.mod"))).MustGet(), nil))
	if !result.IsOk() {
		return result.Error()
	}
	db, _ := lo.Find(meta.SupportedDB(), func(item meta.DBType) bool {
		return item.DB == args[0]
	})
	modules := strings.Join(lo.Without([]string{dbxModule, db.Module}, lo.Map(result.MustGet().Require, func(item *modfile.Require, _ int) string {
		return item.Mod.Path
	})...), " ")
	if _, err := exec.Command("go", "get", "-u", modules).CombinedOutput(); err != nil {
		color.Yellow("failed to install module %s", modules)
	}
	data, _ := tempDir.ReadFile("tmpl/application.yaml")
	rt := mo.TupleToResult(template.New("application").Parse(string(data)))
	//@todo might overwrite file
	mo.TupleToResult(os.Create(filepath.Join(app.RootDir(), fmt.Sprintf("%s.yaml", app.DefaultCfgName))))
	rt.MustGet().Execute(mo.TupleToResult(os.Create(filepath.Join(app.RootDir(), "application.yaml"))).MustGet(), db)
	return nil
}

var initDbCmd = &cobra.Command{
	Use:       "db",
	Short:     "init database configuration for project",
	Long:      "init database configuration for project",
	ValidArgs: []string{"mysql", "pg"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:      initDB,
}

func init() {
	initCmd.AddCommand(initDbCmd)
}
