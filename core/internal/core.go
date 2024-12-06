package internal

import (
	_ "embed"
	"fmt"
	"github.com/kcmvp/gob/core/env"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"log"
	"runtime"
	"strings"
	"sync"
)

var (
	cfg        *viper.Viper
	scope      *do.RootScope
	onceScope  sync.Once
	onceConfig sync.Once
)

func Cfg() *viper.Viper {
	if cfg == nil {
		onceConfig.Do(func() {
			cfg = viper.New()
			cfg.SetConfigName(env.DefaultCfg)          // name of cfg file (without extension)
			cfg.SetConfigType("yaml")                  // REQUIRED if the cfg file does not have the extension in the name
			cfg.AddConfigPath(env.RootDir())           // optionally look for cfg in the working directory
			if err := cfg.ReadInConfig(); err != nil { // Find and read the cfg file
				panic(fmt.Errorf("fatal error cfg file: %w", err))
			}
			if test, _ := Caller(); test {
				tCfg := viper.New()
				tCfg.SetConfigName(fmt.Sprintf("%s_test.yaml", env.DefaultCfg)) // name of cfg file (without extension)
				tCfg.SetConfigType("yaml")                                      // REQUIRED if the cfg file does not have the extension in the name
				tCfg.AddConfigPath(env.RootDir())                               // optionally look for cfg in the working directory
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
		})
	}
	return cfg
}

func Scope() *do.RootScope {
	if scope == nil {
		onceScope.Do(func() {
			// init scope
			scope = do.NewWithOpts(&do.InjectorOpts{
				HookAfterRegistration: []func(scope *do.Scope, serviceName string){
					func(scope *do.Scope, serviceName string) {
						fmt.Printf("scope is %s, name is %s \n", scope.Name(), serviceName)
					},
				},
				Logf: func(format string, args ...any) {
					log.Printf(format, args...)
				},
			}, dbs()...)
		})
	}
	return scope
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
		if strings.HasPrefix(frame.File, env.RootDir()) {
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
