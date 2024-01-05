package internal

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const pluginKey = "plugins"

var (
	once    sync.Once
	project Project
	module  string
)

var PluginExists = PluginExistsError{"plugin exists"}

type PluginExistsError struct {
	errorString string
}

func (p PluginExistsError) Error() string {
	return p.errorString
}

type Project struct {
	root   string
	module string
	deps   []string
	viper  *viper.Viper
	cfg    string
}

func TestCallee() (bool, string) {
	var test bool
	var file string
	callers := make([]uintptr, 10)
	n := runtime.Callers(0, callers)
	frames := runtime.CallersFrames(callers[:n])
	for {
		frame, more := frames.Next()
		// fmt.Printf("%s->%s:%d\n", frame.File, frame.Function, frame.Line)
		test = strings.HasSuffix(frame.File, "_test.go") && strings.HasPrefix(frame.Function, module)
		if test || !more {
			items := strings.Split(frame.File, "/")
			items = lo.Map(items[len(items)-2:], func(item string, _ int) string {
				return strings.ReplaceAll(item, ".go", "")
			})
			file = strings.Join(items, "-")
			break
		}
	}
	return test, file
}

func (project *Project) HookDir() string {
	if ok, file := TestCallee(); ok {
		mock := filepath.Join(CurProject().Target(), file)
		if _, err := os.Stat(mock); err != nil {
			os.Mkdir(mock, os.ModePerm)
		}
		return mock
	} else {
		return filepath.Join(CurProject().Root(), ".git", "hooks")
	}
}

func (project *Project) LoadSettings() {
	testEnv, file := TestCallee()
	// fmt.Printf("caller %s \n", file)
	v := viper.New()
	v.SetConfigType("yaml")
	path := project.Root()
	name := "gob"
	if testEnv {
		path = filepath.Join(project.Target(), file)
		if _, err := os.Stat(path); err != nil {
			if err = os.Mkdir(path, os.ModePerm); err != nil {
				color.Red("failed to create temporary directory %s", path)
			}
		}
	}
	v.AddConfigPath(path)
	v.SetConfigName(name)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			color.Yellow("Warning: can not find configuration gob.yaml")
		}
	}
	project.cfg = fmt.Sprintf("%s.yaml", filepath.Join(path, name))
	project.viper = v
}

func init() {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil || len(string(output)) == 0 {
		log.Fatal(color.RedString("Error: please execute command in project root directory"))
	}

	item := strings.Split(strings.TrimSpace(string(output)), ":")
	//root = item[0]
	module = item[1]
	project = Project{
		root:   item[0],
		module: module,
	}
	cmd = exec.Command("go", "list", "-f", "{{if not .Standard}}{{.ImportPath}}{{end}}", "-deps")
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
	once.Do(func() {
		project.LoadSettings()
	})
	return &project
}

// Configuration gob configuration file
func (project *Project) Configuration() string {
	return project.cfg
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
	target := filepath.Join(project.root, "target")
	if _, err := os.Stat(target); err != nil {
		os.Mkdir(target, os.ModePerm)
	}
	return target
}

// FindGoFilesByPkg return all go source file in a package
func FindGoFilesByPkg(pkg string) ([]string, error) {
	cmd := exec.Command("go", "list", "-f", fmt.Sprintf("{{if eq .Name \"%s\"}}{{.Dir}}{{end}}", pkg), "./...")
	output, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	var dirs []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) > 0 {
			dirs = append(dirs, line)
		}
	}
	if err = scanner.Err(); err != nil {
		return []string{}, err
	}
	return dirs, nil
}

func (project *Project) Plugins() []Plugin {
	if v := project.viper.Get(pluginKey); v != nil {
		plugins := v.(map[string]any)
		return lo.MapToSlice(plugins, func(key string, _ any) Plugin {
			var plugin Plugin
			key = fmt.Sprintf("%s.%s", pluginKey, key)
			if err := project.viper.UnmarshalKey(key, &plugin); err != nil {
				color.Yellow("failed to parse plugin %s: %s", key, err.Error())
			}
			if err := plugin.init(); err != nil {
				color.Red("failed to init plugin %s", plugin.name)
			}
			return plugin
		})
	} else {
		return []Plugin{}
	}
}

func (project *Project) SetupPlugin(plugin Plugin) {
	if !project.isSetup(plugin) {
		if plugin.setup() != nil {
			color.Red("failed to setup plugin %s", plugin.name)
		}
	}
	if plugin.install() != nil {
		color.Red("failed to install plugin %s", plugin.name)
	}
}

func (project *Project) isSetup(plugin Plugin) bool {
	return lo.ContainsBy(project.Plugins(), func(item Plugin) bool {
		return plugin.Module() == item.Module()
	})
}

func (project *Project) Validate() error {
	project.SetupHooks(false)
	lo.ForEach(project.Plugins(), func(plugin Plugin, _ int) {
		plugin.install() //nolint
	})
	return nil
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
		if output, err = exec.Command("git", "show", "--format=%cd", "--date=format:%Y/%m/%d", hash).CombinedOutput(); err == nil {
			return fmt.Sprintf("%s@%s", tag, strings.ReplaceAll(string(output), "\n", ""))
		}
	}
	return unknownVersion
}

func TemporaryGoPath() string {
	dir, _ := os.MkdirTemp("", "test")
	return dir
}

func GoPath() string {
	if ok, method := TestCallee(); ok {
		dir := filepath.Join(os.TempDir(), method, "bin")
		os.MkdirAll(dir, 0o700) //nolint
		return dir
	}
	return filepath.Join(os.Getenv("GOPATH"), "bin")
}

// Windows return true when current os is Windows
func Windows() bool {
	return runtime.GOOS == "windows"
}
