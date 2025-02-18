package meta

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/kcmvp/app"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/spf13/viper"
	"regexp"
)

type GoType string

var (
	//go:embed db.json
	dbJson []byte
)

type DBType struct {
	DB      string `json:"db"`
	Driver  string `json:"driver"`
	Auto    string `json:"auto"`
	Module  string `json:"module"`
	Url     string `json:"url"`
	Default bool   `json:"default"`
}

func SupportedDB() []DBType {
	var supported []DBType
	json.Unmarshal(dbJson, &supported)
	return supported
}

func Platforms() []string {
	var dbs []string
	config := []string{app.DefaultCfgName, fmt.Sprintf("%s_test", app.DefaultCfgName)}
	reg := regexp.MustCompile(`datasource.*\.db`)
	for _, cfgName := range config {
		cfg := viper.New()
		cfg.SetConfigName(cfgName)
		cfg.SetConfigType("yaml")
		cfg.AddConfigPath(app.RootDir())
		if err := cfg.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error cfg file: %w", err))
		}
		dbs = append(dbs, lo.FilterMap(cfg.AllKeys(), func(key string, _ int) (string, bool) {
			if reg.MatchString(key) {
				return cfg.GetString(key), true
			}
			return "", false
		})...)
	}
	return lo.Uniq(dbs)
}

func DB(db string) mo.Option[DBType] {
	return mo.TupleToOption(lo.Find(SupportedDB(), func(dbType DBType) bool {
		return dbType.DB == db
	}))
}

type TypeMapping lo.Tuple4[GoType, string, string, string]

func TypeMappings() []TypeMapping {
	return []TypeMapping{
		{
			A: "string",
			B: "varchar(25)", //MySQL
			C: "varchar(25)", //PostgreSQL
			D: "text(25)",    //SQLite
		},
		{
			A: "bool",
			B: "boolean",
			C: "boolean",
			D: "integer",
		},
		{
			A: "int8",
			B: "tinyint",
			C: "int8",
			D: "integer",
		},
		// unsigned, not a SQL standard
		{
			A: "uint8",
			B: "tinyint unsigned",
			C: "int8",
			D: "integer",
		},
		// unsigned, not a SQL standard
		{
			A: "byte",
			B: "tinyint unsigned",
			C: "int8",
			D: "integer",
		},
		{
			A: "int16",
			B: "smallint",
			C: "smallint",
			D: "integer",
		},
		// unsigned, not a SQL standard
		{
			A: "uint16",
			B: "smallint",
			C: "smallint",
			D: "integer",
		},
		{
			A: "int32",
			B: "int",
			C: "integer",
			D: "integer",
		},
		{
			A: "rune",
			B: "int",
			C: "integer",
			D: "integer",
		},
		// unsigned, not a SQL standard
		{
			A: "uint32",
			B: "int",
			C: "integer",
			D: "integer",
		},
		{
			A: "int64",
			B: "bigint",
			C: "bigint",
			D: "integer",
		},
		{
			A: "int",
			B: "bigint",
			C: "bigint",
			D: "integer",
		},
		// unsigned, not a SQL standard
		{
			A: "uint64",
			B: "bigint",
			C: "bigint",
			D: "integer",
		},
		{
			A: "float32",
			B: "decimal(38, 2)",
			C: "decimal(38, 2)",
			D: "decimal(38, 2)",
		},
		{
			A: "float64",
			B: "decimal(38, 2)",
			C: "decimal(38, 2)",
			D: "decimal(38, 2)",
		},
		{
			A: "time.Time",
			B: "timestamp",
			C: "timestamp",
			D: "datetime",
		},
	}
}
