package internal

import (
	"database/sql"
	"fmt"
	"github.com/kcmvp/gob/core/env"
	"github.com/samber/do/v2"
	typetostring "github.com/samber/go-type-to-string"
	"os"
	"path/filepath"
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
}

func (ds dataSource) DSN() string {
	dsn := strings.ReplaceAll(ds.URL, UserKey, ds.User)
	dsn = strings.ReplaceAll(dsn, PasswordKey, ds.Password)
	return strings.ReplaceAll(dsn, HostKey, ds.Host)
}

func dsMap() map[string]dataSource {
	// single data source
	if v := Cfg().Get(fmt.Sprintf("%s.%s", DSKey, "driver")); v != nil {
		var ds dataSource
		if err := Cfg().UnmarshalKey(DSKey, &ds); err != nil {
			panic(fmt.Errorf("failed parse datasource: %w", err))
		}
		return map[string]dataSource{DefaultDS: ds}
		// multiple data sources
	} else if v = Cfg().Get(DSKey); v != nil {
		dss := v.(map[string]any)
		return lo.MapValues(dss, func(_ any, key string) dataSource {
			var ds dataSource
			key = fmt.Sprintf("%s.%s", DSKey, key)
			if err := Cfg().UnmarshalKey(key, &ds); err != nil {
				panic(fmt.Errorf("failed parse datasource: %w", err))
			}
			return ds
		})
	}
	return map[string]dataSource{}
}

func dbs() []func(do.Injector) {
	var srvs []func(do.Injector)
	for name, ds := range dsMap() {
		if db, err := sql.Open(ds.Driver, ds.DSN()); err == nil {
			if err = db.Ping(); err != nil {
				_ = db.Close()
				panic(fmt.Errorf("failed to initialize %s: %w", name, err))
			}
			lo.ForEach(ds.Scripts, func(script string, _ int) {
				if data, err := os.ReadFile(filepath.Join(env.RootDir(), script)); err == nil {
					if _, err = db.Exec(string(data)); err != nil {
						panic(fmt.Errorf("failed to execute %s: %w", script, err))
					}
				} else {
					panic(fmt.Errorf("failed to read %s: %w", script, err))
				}

			})
			srvs = append(srvs, do.EagerNamed(fmt.Sprintf("%s_%s", name, typetostring.GetType[*sql.DB]()), db))
		}
	}
	return srvs
}
