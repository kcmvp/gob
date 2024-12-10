package internal

import (
	"fmt"
	"github.com/kcmvp/gob/core"
	"strings"

	"github.com/samber/lo"
)

const (
	DSKey       = "datasource"
	UserKey     = "${user}"
	PasswordKey = "${password}"
	HostKey     = "${host}"
	DefaultDS   = "DefaultDS"
)

type dataSource struct {
	DB       string   `mapstructure:"db"`
	Driver   string   `mapstructure:"driver"`
	User     string   `mapstructure:"user"`
	Password string   `mapstructure:"password"`
	Host     string   `mapstructure:"host"`
	URL      string   `mapstructure:"url"`
	Scripts  []string `mapstructure:"scripts"`
	InitDB   bool     `mapstructure:"initDB"` // init database or not, default is false
}

func (ds dataSource) DSN() string {
	dsn := strings.ReplaceAll(ds.URL, UserKey, ds.User)
	dsn = strings.ReplaceAll(dsn, PasswordKey, ds.Password)
	return strings.ReplaceAll(dsn, HostKey, ds.Host)
}

func DSMap() map[string]dataSource {
	// single data source
	cfg := core.Cfg()
	if v := cfg.Get(fmt.Sprintf("%s.%s", DSKey, "driver")); v != nil {
		var ds dataSource
		if err := cfg.UnmarshalKey(DSKey, &ds); err != nil {
			panic(fmt.Errorf("failed parse datasource: %w", err))
		}
		return map[string]dataSource{DefaultDS: ds}
		// multiple data sources
	} else if v = cfg.Get(DSKey); v != nil {
		dss := v.(map[string]any)
		return lo.MapValues(dss, func(_ any, key string) dataSource {
			var ds dataSource
			key = fmt.Sprintf("%s.%s", DSKey, key)
			if err := cfg.UnmarshalKey(key, &ds); err != nil {
				panic(fmt.Errorf("failed parse datasource: %w", err))
			}
			return ds
		})
	}
	return map[string]dataSource{}
}
