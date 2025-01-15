package project

import (
	"embed"
	"errors"
	"github.com/fatih/color" //nolint
	"github.com/kcmvp/gob/common"
	"github.com/samber/lo" //nolint
	"github.com/samber/mo"
	"github.com/spf13/viper" //nolint
	"go/types"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	pluginsKey    = "plugins"
	defaultCfgKey = "_default_"
	buildCfg      = "build.yaml"
)

var (
	latestHash = []string{`log`, `-1`, `--abbrev-commit`, `--date=format-local:%Y-%m-%d %H:%M`, `--format=%h(%ad)`}
	//go:embed resources/*
	resources embed.FS
)

type Project struct {
	root      string
	mod       *modfile.File
	cfgs      sync.Map // store all the configuration
	pkgs      []*packages.Package
	cachDir   string
	module    string
	workspace string
}

func NewProject(root, module string) *Project {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
	mod, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		log.Fatal(color.RedString("failed to read go.mod file %w", err))
	}
	project := &Project{root: root, module: module, mod: mod}
	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  project.root,
	}
	project.pkgs, err = packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(color.RedString("failed to load project %s", err.Error()))
	}
	homeDir, _ := os.UserHomeDir()
	project.cachDir = filepath.Join(homeDir, ".gob", project.Module())
	_ = os.MkdirAll(project.cachDir, os.ModePerm)
	return project
}
func (project *Project) SetWorkSpace(workspace string) {
	project.workspace = workspace
}

func (project *Project) builder() *viper.Viper {
	profile := common.ActiveProfile()
	key := lo.If(profile.Test(), profile.Name()).Else(defaultCfgKey)
	obj, ok := project.cfgs.Load(key)
	if ok {
		return obj.(*viper.Viper)
	}
	v := viper.New()
	path := lo.If(profile.Test(), project.RootDir()).Else(project.TargetDir())
	v.SetConfigFile(filepath.Join(path, buildCfg))
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Fatal(color.RedString("error: can not find build.yaml"))
		}
	}
	project.cfgs.Store(key, v)
	return v
}

func (project *Project) updateCfg(cfg map[string]any) error {
	v := project.builder()
	if err := v.MergeConfigMap(cfg); err != nil {
		return err
	}
	if err := v.WriteConfigAs(v.ConfigFileUsed()); err != nil {
		return err
	}
	_ = v.ReadInConfig()
	prettyYaml(v.ConfigFileUsed())
	return nil
}

//func init() {
//	rs := mo.TupleToResult(exec.Command("go", "list", "-m", "-f", "{{.Dir}}_:_{{.Path}}").CombinedOutput())
//	if rs.IsError() {
//		log.Fatal(color.RedString("can not find go.mod in current directory or any parent directory"))
//	}
//	scanner := bufio.NewScanner(bytes.NewBuffer(rs.MustGet()))
//	var latest string
//	for scanner.Scan() {
//		latest = scanner.Text()
//	}
//	if len(latest) == 0 {
//		log.Fatal(color.RedString("please execute command in project root directory %s", string(latest)))
//	}
//	item := strings.Split(strings.TrimSpace(string(latest)), "_:_")
//	project = Project{cfgs: sync.Map{}, root: item[0]}
//	data, err := os.ReadFile(filepath.Join(project.root, "go.mod"))
//	if err != nil {
//		log.Fatal(color.RedString(err.Error()))
//	}
//	project.mod, err = modfile.Parse("go.mod", data, nil)
//	if err != nil {
//		log.Fatal(color.RedString("please execute command in project root directory %s", string(latest)))
//	}
//	cfg := &packages.Config{
//		Mode: packages.NeedName | packages.NeedTypes | packages.NeedFiles | packages.NeedTypesInfo,
//		Dir:  project.root,
//	}
//	project.pkgs, err = packages.Load(cfg, "./...")
//	if err != nil {
//		log.Fatal(color.RedString("failed to load project %s", err.Error()))
//	}
//	homeDir, err := os.UserHomeDir()
//	if err != nil {
//		return
//	}
//	project.cachDir = filepath.Join(homeDir, ".gob", Module())
//	_ = os.MkdirAll(project.cachDir, os.ModePerm)
//}

func (project *Project) CacheDir() string {
	return project.cachDir
}

// RootDir return root dir of the project
func (project *Project) RootDir() string {
	return project.root
}

// Module return current project module name
func (project *Project) Module() string {
	return project.module
}

func (project *Project) TargetDir() string {
	target := filepath.Join(project.RootDir(), "target")
	if profile := common.ActiveProfile(); profile.Test() {
		target = filepath.Join(target, profile.Name())
	}
	if rs := mo.TupleToResult(os.Stat(target)); rs.IsError() {
		_ = os.MkdirAll(target, os.ModePerm)
	}
	return target
}

func (project *Project) MainFiles() []string {
	return lo.FilterMap(project.pkgs, func(pkg *packages.Package, _ int) (string, bool) {
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
}

//func (project *Project) Plugins() []Plugin {
//	viper := project.builder()
//	if v := viper.Get(pluginsKey); v != nil {
//		config := v.(map[string]any)
//		plugins := lo.MapToSlice(config, func(key string, _ any) Plugin {
//			var plugin Plugin
//			if err := viper.UnmarshalKey(fmt.Sprintf("%s.%s", pluginsKey, key), &plugin); err != nil {
//				color.Yellow("failed to parse plugin %s: %s", key, err.Error())
//			}
//			if len(strings.TrimSpace(key)) == 0 {
//				color.Yellow("empty plugin name %s", plugin.Url)
//			}
//			plugin.Name = key
//			return plugin
//		})
//		return lo.Filter(plugins, func(p Plugin, _ int) bool {
//			return len(strings.TrimSpace(p.Name)) > 0
//		})
//	} else {
//		return []Plugin{}
//	}
//}

func InGit() bool {
	_, err := exec.Command("git", "status").CombinedOutput()
	return err == nil
}

func Version() string {
	version := "unknown"
	if output, err := exec.Command("git", latestHash...).CombinedOutput(); err == nil {
		version = strings.Trim(string(output), "\n")
	}
	return version
}

func temporaryGoPath() string {
	dir, _ := os.MkdirTemp("", "gob-build-")
	return dir
}

func GoPath() string {
	if profile := common.ActiveProfile(); profile.Test() {
		dir := filepath.Join(os.TempDir(), profile.Name())
		_ = os.MkdirAll(dir, os.ModePerm) //nolint
		return dir
	}
	return filepath.Join(os.Getenv("GOPATH"), "bin")
}
