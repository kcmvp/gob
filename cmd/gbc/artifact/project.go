package artifact

import (
	"errors"
	"fmt"
	"github.com/fatih/color" //nolint
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"   //nolint
	"github.com/spf13/viper" //nolint
	"go/types"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	pluginCfgKey  = "plugins"
	defaultCfgKey = "_default_"
	pluginCfgFile = "plugins.yaml"
)

var (
	project Project
)

type Project struct {
	root string
	mod  *modfile.File
	cfgs sync.Map // store all the configuration
	pkgs []*packages.Package
}

func (project *Project) load() *viper.Viper {
	testEnv, file := utils.TestCaller()
	key := lo.If(testEnv, file).Else(defaultCfgKey)
	obj, ok := project.cfgs.Load(key)
	if ok {
		return obj.(*viper.Viper)
	}
	v := viper.New()
	path := lo.If(!testEnv, project.Root()).Else(project.Target())
	v.SetConfigFile(filepath.Join(path, "gob.yaml"))
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Fatal(color.RedString("error: can not find configuration gob.yaml"))
		}
	}
	project.cfgs.Store(key, v)
	return v
}

func (project *Project) mergeConfig(cfg map[string]any) error {
	viper := project.load()
	err := viper.MergeConfigMap(cfg)
	if err != nil {
		return err
	}
	return viper.WriteConfigAs(viper.ConfigFileUsed())
}

func (project *Project) HookDir() string {
	if ok, _ := utils.TestCaller(); ok {
		mock := filepath.Join(CurProject().Target(), ".git", "hooks")
		if _, err := os.Stat(mock); err != nil {
			os.MkdirAll(mock, os.ModePerm) //nolint
		}
		return mock
	}
	return filepath.Join(CurProject().Root(), ".git", "hooks")
}

func init() {
	output, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}_:_{{.Path}}").CombinedOutput()
	if err != nil || len(string(output)) == 0 {
		log.Fatal(color.RedString("please execute command in project root directory %s", string(output)))
	}
	item := strings.Split(strings.TrimSpace(string(output)), "_:_")
	project = Project{cfgs: sync.Map{}, root: item[0]}
	data, err := os.ReadFile(filepath.Join(project.root, "go.mod"))
	if err != nil {
		log.Fatal(color.RedString(err.Error()))
	}
	project.mod, err = modfile.Parse("go.mod", data, nil)
	if err != nil {
		log.Fatal(color.RedString("please execute command in project root directory %s", string(output)))
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax,
		Dir:  project.root,
	}
	project.pkgs, err = packages.Load(cfg, "./...")
	if err != nil {
		log.Fatal(color.RedString("failed to load project %s", err.Error()))
	}
}

// CurProject return Project struct
func CurProject() *Project {
	return &project
}

// Root return root dir of the project
func (project *Project) Root() string {
	return project.root
}

// Module return current project module name
func (project *Project) Module() string {
	return project.mod.Module.Mod.Path
}

func (project *Project) Target() string {
	target := filepath.Join(project.Root(), "target")
	if test, method := utils.TestCaller(); test {
		target = filepath.Join(target, method)
	}
	if _, err := os.Stat(target); err != nil {
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

func (project *Project) Plugins() []Plugin {
	viper := project.load()
	if v := viper.Get(pluginCfgKey); v != nil {
		plugins := v.(map[string]any)
		return lo.MapToSlice(plugins, func(key string, _ any) Plugin {
			var plugin Plugin
			key = fmt.Sprintf("%s.%s", pluginCfgKey, key)
			if err := viper.UnmarshalKey(key, &plugin); err != nil {
				color.Yellow("failed to parse plugin %s: %s", key, err.Error())
			}
			if err := plugin.init(); err != nil {
				color.Red("failed to init plugin %s: %s", plugin.name, err.Error())
			}
			return plugin
		})
	} else {
		return []Plugin{}
	}
}

func (project *Project) Dependencies() []*modfile.Require {
	return project.mod.Require
}

func (project *Project) InstallDependency(dep string) error {
	var err error
	if lo.NoneBy(project.mod.Require, func(r *modfile.Require) bool {
		return lo.Contains(r.Syntax.Token, dep)
	}) {
		_, err = exec.Command("go", "get", "-u", dep).CombinedOutput() //nolint
	}
	return err
}

func (project *Project) InstallPlugin(plugin Plugin) error {
	var err error
	if !project.settled(plugin) {
		values := lo.MapEntries(map[string]string{
			"alias": plugin.Alias,
			"args":  plugin.Args,
			"url":   fmt.Sprintf("%s@%s", plugin.Url, plugin.Version()),
		}, func(key string, value string) (string, any) {
			return fmt.Sprintf("%s.%s.%s", pluginCfgKey, plugin.Name(), key), value
		})
		if err = project.mergeConfig(values); err != nil {
			return err
		}
		_ = project.load().ReadInConfig()
	}
	_, err = plugin.install()
	return err
}

func (project *Project) settled(plugin Plugin) bool {
	return project.load().Get(fmt.Sprintf("plugins.%s.url", plugin.name)) != nil
}

func (project *Project) Validate() error {
	return project.SetupHooks(false)
}

func InGit() bool {
	_, err := exec.Command("git", "status").CombinedOutput()
	return err == nil
}

var latestHash = []string{`log`, `-1`, `--abbrev-commit`, `--date=format-local:%Y-%m-%d %H:%M`, `--format=%h(%ad)`}

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
	if ok, method := utils.TestCaller(); ok {
		dir := filepath.Join(os.TempDir(), method)
		_ = os.MkdirAll(dir, os.ModePerm) //nolint
		return dir
	}
	return filepath.Join(os.Getenv("GOPATH"), "bin")
}

// Windows return true when current os is Windows
func Windows() bool {
	return runtime.GOOS == "windows"
}
