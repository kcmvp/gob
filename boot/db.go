package boot

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kcmvp/gob/internal"

	//"github.com/kcmvp/gob/internal"

	"github.com/samber/do/v2"
	typetostring "github.com/samber/go-type-to-string"
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

func setupDb() {
	for name, ds := range dsMap() {
		if db, err := sql.Open(ds.Driver, ds.DSN()); err == nil {
			if err = db.Ping(); err != nil {
				_ = db.Close()
				panic(fmt.Errorf("failed to initialize %s: %w", name, err))
			}
			lo.ForEach(ds.Scripts, func(script string, _ int) {
				if data, err := os.ReadFile(filepath.Join(internal.RootDir, script)); err == nil {
					if _, err = db.Exec(string(data)); err != nil {
						panic(fmt.Errorf("failed to execute %s: %w", script, err))
					}
				} else {
					panic(fmt.Errorf("failed to read %s: %w", script, err))
				}
			})
			do.ProvideNamedValue[*sql.DB](internal.Container, fmt.Sprintf("%s_%s", name, typetostring.GetType[*sql.DB]()), db)
		} else {
			panic(fmt.Errorf("failed to connect to datasource %s: %w", name, err))
		}
	}
}
