package internal

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fatih/color" //nolint
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
)

var (
	project Project
)

type Project struct {
	root   string
	module string
	deps   []string
	cfg    sync.Map // store all the configuration
}

// TestCaller returns true when caller is from _test.go and the full method name
func TestCaller() (bool, string) {
	var test bool
	var file string
	callers := make([]uintptr, 10)
	n := runtime.Callers(0, callers)
	frames := runtime.CallersFrames(callers[:n])
	for {
		frame, more := frames.Next()
		// fmt.Printf("%s->%s:%d\n", frame.File, frame.Function, frame.Line)
		test = strings.HasSuffix(frame.File, "_test.go") && strings.HasPrefix(frame.Function, project.module)
		if test || !more {
			items := strings.Split(frame.File, "/")
			items = lo.Map(items[len(items)-2:], func(item string, _ int) string {
				return strings.ReplaceAll(item, ".go", "")
			})
			uniqueNames := strings.Split(frame.Function, ".")
			items = append(items, uniqueNames[len(uniqueNames)-1])
			file = strings.Join(items, "_")
			break
		}
	}
	return test, file
}

func (project *Project) config() *viper.Viper {
	testEnv, file := TestCaller()
	key := lo.If(testEnv, file).Else(defaultCfgKey)
	obj, ok := project.cfg.Load(key)
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
	project.cfg.Store(key, v)
	return v
}

func (project *Project) mergeConfig(cfg map[string]any) error {
	err := project.config().MergeConfigMap(cfg)
	if err != nil {
		return err
	}
	return project.config().WriteConfigAs(project.config().ConfigFileUsed())
}

func (project *Project) HookDir() string {
	if ok, _ := TestCaller(); ok {
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
		cfg:    sync.Map{},
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
	if test, method := TestCaller(); test {
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
	if v := project.config().Get(pluginCfgKey); v != nil {
		plugins := v.(map[string]any)
		return lo.MapToSlice(plugins, func(key string, _ any) Plugin {
			var plugin Plugin
			key = fmt.Sprintf("%s.%s", pluginCfgKey, key)
			if err := project.config().UnmarshalKey(key, &plugin); err != nil {
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

func (project *Project) SetupPlugin(plugin Plugin) {
	if !project.isSetup(plugin) {
		values := lo.MapEntries(map[string]string{
			"alias": plugin.Alias,
			"args":  plugin.Args,
			"url":   fmt.Sprintf("%s@%s", plugin.Url, plugin.Version()),
		}, func(key string, value string) (string, any) {
			return fmt.Sprintf("%s.%s.%s", pluginCfgKey, plugin.Name(), key), value
		})
		if err := project.mergeConfig(values); err != nil {
			color.Red("faialed to setup plugin %s", err.Error())
			return
		}
		_ = project.config().ReadInConfig()
	}
	if _, err := plugin.install(); err != nil {
		color.Red("failed to install plugin %s: %s", plugin.name, err.Error())
	}
}

func (project *Project) isSetup(plugin Plugin) bool {
	return project.config().Get(fmt.Sprintf("plugins.%s.url", plugin.name)) != nil
}

func (project *Project) Validate() {
	project.SetupHooks(false)
}

func InGit() bool {
	_, err := exec.Command("git", "status").CombinedOutput()
	return err == nil
}

var unknownVersion = "unknown"

func Version() string {
	if output, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput(); err == nil {
		hash := strings.ReplaceAll(string(output), "\n", "")
		output, err = exec.Command("git", "describe", "--tag", hash).CombinedOutput()
		if err != nil {
			return unknownVersion
		}
		tag := strings.ReplaceAll(string(output), "\n", "")
		output, _ = exec.Command("git", "status", "--short").CombinedOutput()
		if lo.ContainsBy(strings.Split(string(output), "\n"), func(line string) bool {
			line = strings.TrimSpace(strings.ToUpper(line))
			return strings.HasPrefix(line, "M ") ||
				strings.HasSuffix(line, "D ") ||
				strings.HasSuffix(line, "?? ")
		}) {
			return fmt.Sprintf("%s@stage", tag)
		}
		if output, err = exec.Command("git", "log", "--format=%ci -n 1", hash).CombinedOutput(); err == nil {
			return fmt.Sprintf("%s@%s", tag, strings.ReplaceAll(string(output), "\n", ""))
		}
	}
	return unknownVersion
}

func temporaryGoPath() string {
	dir, _ := os.MkdirTemp("", "gob-build-")
	return dir
}

func GoPath() string {
	if ok, method := TestCaller(); ok {
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
