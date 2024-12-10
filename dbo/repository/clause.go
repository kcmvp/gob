package repository

import (
	"fmt"
	. "github.com/kcmvp/dbo/entity"
	"github.com/samber/lo"
)

type Column[E IEntity] lo.Tuple3[string, string, string]

func Mapping[E IEntity](attr, column, props string) Column[E] {
	return Column[E](lo.Tuple3[string, string, string]{A: attr, B: column, C: props})
}

func (column Column[E]) Sql() string {
	var zero E
	return fmt.Sprintf("%s.%s", zero.Table(), column.B)
}

func (column Column[E]) PK() bool {
	return false
}

type Where[E IEntity] struct {
	zero E
	exp  string
	val  any
}

func None[E IEntity]() Where[E] {
	return Where[E]{
		exp: "",
	}
}

//lo.Tuple2[string, any]

func (criteria Where[E]) eval() (string, []any) {
	panic("implement me")
}

func (column Column[E]) LessThan(v any) Where[E] {
	return Where[E]{
		exp: column.Sql(),
		val: v,
	}
}

func (column Column[E]) LessThanEqual(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s < ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) GreaterThan(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s > ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) GreaterThanEqual(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) IsNull(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) NotNull(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) IsTrue(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) NotTrue(v any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: v,
	}
}

func (column Column[E]) Between(v1, v2 any) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: []any{v1, v2},
	}
}

func (column Column[E]) Like(v string) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: []any{v},
	}
}

func (column Column[E]) NotLike(v string) Where[E] {
	return Where[E]{
		exp: fmt.Sprintf("%s >= ?", column.Sql()),
		val: []any{v},
	}
}

func And[E IEntity](l Where[E], r Where[E]) Where[E] {
	return Where[E]{
		exp: "",
		val: []any{l.val, r.val},
	}
}

func Or[E IEntity](l Where[E], r Where[E]) Where[E] {
	return Where[E]{
		exp: "",
		val: []any{l.val, r.val},
	}
}

type Set[E IEntity] struct {
	zero   E
	column Column[E]
	val    any
}

func (column Column[E]) Set(v any) Set[E] {
	return Set[E]{
		column: column,
		val:    v,
	}
}

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

type OrderBy lo.Tuple2[int, Order]
