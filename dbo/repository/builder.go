package repository

import (
	"database/sql"
	. "github.com/kcmvp/dbo/entity"
	"github.com/samber/lo"
)

type Joint[E1 IEntity, E2 IEntity] lo.Tuple2[Column[E1], Column[E2]]

func (joint Joint[E1, E2]) tables() []string {
	var z1 E1
	var z2 E2
	return []string{z1.Table(), z2.Table()}
}

type Join[E1 IEntity, E2 IEntity] func(Column[E1], Column[E2]) Joint[E1, E2]

func Inner[E1 IEntity, E2 IEntity](c1 Column[E1], c2 Column[E2]) Joint[E1, E2] {
	panic("implement me")
}

func Left[E1 IEntity, E2 IEntity](c1 Column[E1], c2 Column[E2]) Joint[E1, E2] {
	panic("implement me")
}

type SqlBuilder struct {
	joints []Joint[IEntity, IEntity]
	clause string
}

func (builder SqlBuilder) tables() []string {
	var tables []string
	for _, joint := range builder.joints {
		tables = append(joint.tables())
	}
	return tables
}

func (builder SqlBuilder) Select(cols []Column[IEntity], where Where[IEntity], orders ...OrderBy) (*sql.Row, error) {
	panic("implement me")
}

func (builder SqlBuilder) Update(sets []Set[IEntity], where Where[IEntity]) (int64, error) {
	panic("not implemented")
}
