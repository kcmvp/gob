package template

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	viperOnce     sync.Once
	viperInstance *viper.Viper
)

func NewViper() *viper.Viper {
	viperOnce.Do(func() {
		viperInstance = viper.New()
		viper.SetConfigName("application") // name of config file (without extension)
		viper.SetConfigType("yaml")        // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(".")           // optionally look for config in the working directory
		err := viper.ReadInConfig()        // Find and read the config file
		if err != nil {                    // Handle errors reading the config file
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	})
	return viperInstance
}
