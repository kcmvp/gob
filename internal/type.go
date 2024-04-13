package internal

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

const (
	MySQL      = "mysql"
	PostgreSQL = "postgreSQL"
	SQLite     = "sqlite"
)

type (
	GoType string
	DBType lo.Tuple3[string, []string, []string]
)

func ParseDBType() (DBType, error) {
	var dbName string
	db, ok := lo.Find([]DBType{
		{A: MySQL, B: []string{"auto_increment"}, C: []string{"github.com/go-sql-driver/mysql"}},
		{A: PostgreSQL, B: []string{"smallserial", "serial", "bigserial"}, C: []string{"github.com/lib/pq", "github.com/jackc/pgx"}},
		{A: SQLite, B: []string{"autoincrement"}, C: []string{"github.com/mattn/go-sqlite3"}},
	}, func(item DBType) bool {
		return strings.Contains(dbName, item.A)
	})
	if !ok {
		return DBType{}, fmt.Errorf("can not find database driver in go.mod")
	}
	return db, nil
}

func (p DBType) PrimaryStr() []string {
	return p.B
}

var TypeMappings = []lo.Tuple4[GoType, string, string, string]{
	{
		A: "string",
		B: "varchar", // MySQL
		C: "varchar", // PostgreSQL
		D: "text",    // SQLite
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
	// unsigned, not a SQL standard
	{
		A: "uint64",
		B: "bigint",
		C: "bigint",
		D: "integer",
	},
	{
		A: "float32",
		B: "float",
		C: "double precision",
		D: "double precision",
	},
	{
		A: "float64",
		B: "decimal",
		C: "decimal",
		D: "decimal",
	},
	{
		A: "time",
		B: "timestamp",
		C: "timestamp",
		D: "datetime",
	},
}
