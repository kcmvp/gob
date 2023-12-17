package internal

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"io/fs"
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
	//root    string
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

func TestEnv() (bool, string) {
	var test bool
	var file string
	callers := make([]uintptr, 20)
	n := runtime.Callers(0, callers)
	frames := runtime.CallersFrames(callers[:n])
	for {
		frame, more := frames.Next()
		test = strings.HasSuffix(frame.File, "_test.go") && strings.HasPrefix(frame.Function, module)
		if test || !more {
			file, _ = lo.Last(strings.Split(frame.Function, "."))
			break
		}
	}
	return test, file
}

func (project *Project) LoadSettings() {
	testEnv, _ := TestEnv()
	v := viper.New()
	v.SetConfigType("yaml")
	path := project.Root()
	name := "gb"
	if testEnv {
		name = fmt.Sprintf("gb-%s", lo.RandomString(12, lo.AlphanumericCharset))
		path = project.Target()
	}
	v.AddConfigPath(path)
	v.SetConfigName(name)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			color.Yellow("Warning: can not find configuration gb.yaml")
		}
	}
	project.cfg = fmt.Sprintf("%s.yaml", filepath.Join(path, name))
	project.viper = v
}

func init() {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil || len(string(output)) == 0 {
		log.Fatal(color.RedString("Error executing command: %v", err))
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
		log.Fatal(color.RedString("Error executing command: %v", err))
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

func (project *Project) Config() string {
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

// Plugins all the configured plugins
func (project *Project) Plugins() []lo.Tuple4[string, string, string, string] {
	if project.viper.Get(pluginKey) != nil {
		plugins := project.viper.Get(pluginKey).(map[string]any)
		return lo.MapToSlice(plugins, func(key string, value any) lo.Tuple4[string, string, string, string] {
			attr := value.(map[string]any)
			//@todo validate attribute for null or empty
			return lo.Tuple4[string, string, string, string]{
				key, attr["alias"].(string), attr["command"].(string), attr["url"].(string),
			}
		})
	} else {
		return []lo.Tuple4[string, string, string, string]{}
	}
}

// NormalizePlugin returns the base name and versioned name of the plugin
func NormalizePlugin(url string) (base string, name string) {
	name, _ = lo.Last(strings.Split(url, "/"))
	base = strings.Split(name, "@")[0]
	name = strings.ReplaceAll(name, "@", "-")
	if Windows() {
		name = fmt.Sprintf("%s.exe", name)
	}
	return
}

// PluginInstalled return true if the plugin is installed
func (project *Project) PluginInstalled(url string) bool {
	_, name := NormalizePlugin(url)
	_, err := os.Stat(filepath.Join(os.Getenv("GOPATH"), "bin", name))
	return err == nil
}

func (project *Project) PluginConfigured(url string) bool {
	_, ok := lo.Find(CurProject().Plugins(), func(item lo.Tuple4[string, string, string, string]) bool {
		return item.D == strings.TrimSpace(url)
	})
	return ok
}

// InstallPlugin install the tool as gb plugin save it in gb.yml
func (project *Project) InstallPlugin(url string, aliasAndCommand ...string) error {
	base, name := NormalizePlugin(url)
	gopath := os.Getenv("GOPATH")
	installed := project.PluginInstalled(url)
	configured := project.PluginConfigured(url)
	if installed && configured {
		return PluginExists
	} else {
		var err error
		if !installed {
			// install only
			//dir, _ := os.MkdirTemp("", base)
			//os.Setenv("GOPATH", dir)
			fmt.Sprintf("Installing %s ...... \n", url)
			_, err = exec.Command("go", "install", url).CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to install %s: %v", url, err)
			}
			defer func() {
				//os.Setenv("GOPATH", gopath)
				//os.RemoveAll(dir)
			}()
			if err = filepath.WalkDir(gopath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasPrefix(d.Name(), base) {
					err = os.Rename(path, filepath.Join(gopath, "bin", name))
					if err != nil {
						return err
					}
					fmt.Printf("%s is installed successfully \n", url)
					return filepath.SkipAll
				}
				return err
			}); err != nil {
				return err
			}
		}
		if !configured {
			// install & update configuration
			fmt.Println("Updating configuration ......")
			var alias, command string
			if len(aliasAndCommand) > 0 {
				alias = aliasAndCommand[0]
			}
			if len(aliasAndCommand) > 1 {
				command = aliasAndCommand[1]
			}
			project.viper.Set(fmt.Sprintf("%s.%s.%s", pluginKey, base, "alias"), alias)
			project.viper.Set(fmt.Sprintf("%s.%s.%s", pluginKey, base, "command"), command)
			project.viper.Set(fmt.Sprintf("%s.%s.%s", pluginKey, base, "url"), url)
			err = project.viper.WriteConfigAs(project.cfg)
		}
		return err
	}
}

// Windows return true when current os is Windows
func Windows() bool {
	return runtime.GOOS == "windows"
}
