package core

import (
	_ "embed"
	"fmt"
	"github.com/kcmvp/gob/core/utils"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

var (
	once    sync.Once
	app     *Application
	rootDir string
	cfg     *viper.Viper
)

const (
	DefaultCfg = "application"
)

type Application struct {
	do.Injector
	cfg  *viper.Viper
	root string
}

func init() {
	// get project root dir
	dir, _ := exec.Command("go", "list", "-m", "-f", "{{.Dir}}").CombinedOutput()
	rootDir = utils.CleanStr(string(dir))
	if len(rootDir) == 0 {
		rootDir = mo.TupleToResult(os.Executable()).MustGet()
	}
	// init project config
	cfg = viper.New()
	cfg.SetConfigName(DefaultCfg)              // name of cfg file (without extension)
	cfg.SetConfigType("yaml")                  // REQUIRED if the cfg file does not have the extension in the name
	cfg.AddConfigPath(rootDir)                 // optionally look for cfg in the working directory
	if err := cfg.ReadInConfig(); err != nil { // Find and read the cfg file
		panic(fmt.Errorf("fatal error cfg file: %w", err))
	}
	// merge the configuration
	// @todo need to support profile environment
	if test, _ := Caller(); test {
		tCfg := viper.New()
		tCfg.SetConfigName(fmt.Sprintf("%s_test.yaml", DefaultCfg)) // name of cfg file (without extension)
		tCfg.SetConfigType("yaml")                                  // REQUIRED if the cfg file does not have the extension in the name
		tCfg.AddConfigPath(rootDir)                                 // optionally look for cfg in the working directory
		if err := tCfg.ReadInConfig(); err != nil {
			panic(fmt.Errorf("failed to merge test configuration file: %w", err))
		}
		rootKeys := lo.Uniq(lo.Map(tCfg.AllKeys(), func(key string, index int) string {
			return strings.Split(key, ".")[0]
		}))
		patch := map[string]any{}
		lo.ForEach(cfg.AllKeys(), func(key string, _ int) {
			if lo.NoneBy(rootKeys, func(root string) bool {
				return strings.HasPrefix(key, root)
			}) {
				patch[key] = cfg.Get(key)
			}
		})
		if err := tCfg.MergeConfigMap(patch); err != nil {
			panic(fmt.Errorf("failed to merge test configuration file: %w", err))
		}
		cfg = tCfg
	}
}

func Caller() (bool, string) {
	var test bool
	var file string
	callers := make([]uintptr, 50)
	n := runtime.Callers(0, callers)
	frames := runtime.CallersFrames(callers[:n])
	for !test {
		frame, more := frames.Next()
		if !more {
			break
		}
		// fmt.Printf("%s->%s:%d\n", frame.File, frame.Function, frame.Line)
		if strings.HasPrefix(frame.File, rootDir) {
			test = strings.HasSuffix(frame.File, "_test.go")
			items := strings.Split(frame.File, "/")
			items = lo.Map(items[len(items)-2:], func(item string, _ int) string {
				return strings.ReplaceAll(item, ".go", "")
			})
			uniqueNames := strings.Split(frame.Function, ".")
			items = append(items, uniqueNames[len(uniqueNames)-1])
			file = strings.Join(items, "_")
		}
	}
	return test, file
}

func App() *Application {
	if app == nil {
		once.Do(func() {
			app = &Application{
				Injector: do.NewWithOpts(&do.InjectorOpts{
					HookAfterRegistration: []func(scope *do.Scope, serviceName string){
						func(scope *do.Scope, serviceName string) {
							fmt.Printf("scope is %s, name is %s \n", scope.Name(), serviceName)
						},
					},
					Logf: func(format string, args ...any) {
						log.Printf(format, args...)
					},
				}),
				cfg:  cfg,
				root: rootDir,
			}
		})
	}
	return app
}

func (app *Application) Cfg() *viper.Viper {
	return app.cfg
}
func Cfg() *viper.Viper {
	return App().cfg
}

func (app *Application) RootDir() string {
	return app.root
}
func RootDir() string {
	return App().root
}

type ContextAware func(*viper.Viper) func(do.Injector)

func Context(service ContextAware) func(do.Injector) {
	return service(App().cfg)
}

func Register(servers ...func(do.Injector)) {
	do.Package(servers...)(App())
}
