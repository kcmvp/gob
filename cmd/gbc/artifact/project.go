package artifact

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fatih/color" //nolint
	"github.com/kcmvp/gob/utils"
	"github.com/samber/lo"   //nolint
	"github.com/spf13/viper" //nolint
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
	root   string
	module string
	deps   []string
	cfgs   sync.Map // store all the configuration
}

func (project *Project) config() *viper.Viper {
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
			color.Yellow("Warning: can not find configuration gob.yaml")
		}
	}
	project.cfgs.Store(key, v)
	return v
}

func (project *Project) mergeConfig(cfg map[string]any) error {
	viper := project.config()
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
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil || len(string(output)) == 0 {
		log.Fatal(color.RedString("Error: please execute command in project root directory"))
	}

	item := strings.Split(strings.TrimSpace(string(output)), ":")
	project = Project{
		root:   item[0],
		module: item[1],
		cfgs:   sync.Map{},
	}
	cmd = exec.Command("go", "list", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "-deps", "./...")
	output, err = cmd.Output()
	if err != nil {
		log.Fatal(color.RedString("Error: please execute command in project root directory"))
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var deps []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			deps = append(deps, line)
		}
	}
	project.deps = deps
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
	return project.module
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

// sourceFileInPkg return all go source file in a package
func (project *Project) sourceFileInPkg(pkg string) ([]string, error) {
	_ = os.Chdir(project.Root())
	cmd := exec.Command("go", "list", "-f", fmt.Sprintf("{{if eq .Name \"%s\"}}{{.Dir}}{{end}}", pkg), "./...")
	output, _ := cmd.Output()
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var dirs []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			dirs = append(dirs, line)
		}
	}
	return dirs, nil
}

func (project *Project) MainFiles() []string {
	var mainFiles []string
	dirs, _ := project.sourceFileInPkg("main")
	re := regexp.MustCompile(`func\s+main\s*\(\s*\)`)
	lo.ForEach(dirs, func(dir string, _ int) {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() && dir != path {
				return filepath.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if re.MatchString(line) {
					mainFiles = append(mainFiles, path)
					return filepath.SkipDir
				}
			}
			return scanner.Err()
		})
	})
	return mainFiles
}

func (project *Project) Plugins() []Plugin {
	viper := project.config()
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

func (project *Project) Dependencies() []string {
	return project.deps
}

func (project *Project) InstallDependency(dep string) error {
	if !lo.Contains(project.deps, dep) {
		exec.Command("go", "get", "-u", dep).CombinedOutput() //nolint
	}
	return nil
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
		_ = project.config().ReadInConfig()
	}
	_, err = plugin.install()
	return err
}

func (project *Project) settled(plugin Plugin) bool {
	return project.config().Get(fmt.Sprintf("plugins.%s.url", plugin.name)) != nil
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
