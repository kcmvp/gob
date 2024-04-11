package internal

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/kcmvp/structs"
	"github.com/samber/lo"
)

type Mapper lo.Tuple3[string, string, []string]

func (mapper Mapper) String() string {
	return fmt.Sprintf("%s_%s_%s", mapper.A, mapper.B, strings.Join(mapper.C, "_"))
}

func (mapper Mapper) IsPK() bool {
	return lo.ContainsBy(mapper.C, func(item string) bool {
		return string(PK) == item
	})
}

type Attribute string

const (
	// DBTag struct tag name
	DBTag = "db"
	// AutoUpdateTime attribute tag with `aut` will be set to time.Now() for update
	AutoUpdateTime Attribute = "aut"
	// AutoCreateTime attribute tag with `act` will be set to time.Now() for creation
	AutoCreateTime Attribute = "act"
	// PK identify a column is primary key
	PK Attribute = "pk"
	// Ignore identify this attribute would not map to database column
	Ignore Attribute = "ignore"
	// Name database column name
	Name Attribute = "name"
	// Type database column type
	Type Attribute = "type"
)

func Parse(str any) []Mapper {
	var mappers []Mapper
	return append(mappers, lo.FilterMap(structs.Fields(str), func(f *structs.Field, _ int) (Mapper, bool) {
		if !f.IsExported() {
			return Mapper{}, false
		}
		if f.IsEmbedded() && f.Kind().String() == "struct" {
			mappers = append(mappers, Parse(f.Value())...)
			return Mapper{}, false
		}
		colName := strcase.ToSnake(f.Name())
		var attrs []string
		ignore := false
		if tag := f.Tag(DBTag); tag != "" {
			prefix := fmt.Sprintf("%s=", Name)
			attrs = strings.Split(tag, ";")
			ignore = lo.ContainsBy(attrs, func(attr string) bool {
				return string(Ignore) == strings.TrimSpace(attr)
			})
			if !ignore {
				lo.ForEach(attrs, func(item string, _ int) {
					if strings.HasPrefix(item, prefix) {
						colName = strings.TrimLeft(item, prefix)
					}
				})
			}
		}
		return Mapper{A: f.Name(), B: colName, C: attrs}, !ignore
	})...)
}
