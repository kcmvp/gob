package boot

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kcmvp/gob/internal"
	"github.com/kcmvp/gob/utils"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

const (
	DefaultCfg = "application"
)

var (
	cfg  *viper.Viper
	once sync.Once
)

func RootDir() string {
	return internal.RootDir
}

func Container() *do.RootScope {
	return internal.Container
}

func InitApp() {
	InitAppWith(DefaultCfg)
}

func InitAppWith(cfgName string) {
	if cfg == nil {
		once.Do(func() {
			cfg = viper.New()
			cfg.SetConfigName(cfgName)                 // name of cfg file (without extension)
			cfg.SetConfigType("yaml")                  // REQUIRED if the cfg file does not have the extension in the name
			cfg.AddConfigPath(internal.RootDir)        // optionally look for cfg in the working directory
			if err := cfg.ReadInConfig(); err != nil { // Find and read the cfg file
				panic(fmt.Errorf("fatal error cfg file: %w", err))
			}
			if test, _ := utils.TestCaller(); test {
				if testCfg, err := os.Open(filepath.Join(internal.RootDir, fmt.Sprintf("%s_test.yaml", cfgName))); err == nil {
					if err = cfg.MergeConfig(testCfg); err != nil {
						panic(fmt.Errorf("failed to merge test configuration file: %w", err))
					}
				}
			}
			setupDb()
		})
	}
}
