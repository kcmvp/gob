package dbx

import (
	"fmt"

	"github.com/kcmvp/gob/internal"
	"github.com/samber/lo"
)

type IEntity interface {
	Table() string
}

type Key interface {
	string | int64
}

type SQL interface {
	SQLStr() string
}

type Order string

const (
	ASC  Order = "ASC"
	DESC Order = "DESC"
)

type Attr internal.Mapper

func (a Attr) SQLStr() string {
	return a.B
}

type Set lo.Tuple2[Attr, any]

// Criteria attribute predicate
type Criteria struct {
	expression string
}

func (criteria Criteria) SQLStr() string {
	return criteria.expression
}

func (criteria Criteria) Or(rp Criteria) Criteria {
	return Criteria{
		expression: fmt.Sprintf("(%s or %s)", criteria.expression, rp.SQLStr()),
	}
}

func (criteria Criteria) Add(rp Criteria) Criteria {
	return Criteria{
		expression: fmt.Sprintf("(%s and %s)", criteria.expression, rp.SQLStr()),
	}
}

func LT(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s < ?", attr.SQLStr())}
}

func LTE(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s <= ?", attr.SQLStr())}
}

func GT(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s > ?", attr.SQLStr())}
}

func GTE(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s >= ?", attr.SQLStr())}
}

func Null(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s is null", attr.SQLStr())}
}

func NotNull(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s is not null", attr.SQLStr())}
}

func Like(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s like '%%?%%'", attr.SQLStr())}
}

func NotLike(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s not like '%%?%%'", attr.SQLStr())}
}

func Prefix(attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s like '?%%'", attr.SQLStr())}
}

func Suffix[E IEntity](attr Attr) Criteria {
	return Criteria{expression: fmt.Sprintf("%s like '%%?'", attr.SQLStr())}
}

type OrderBy lo.Tuple2[Attr, Order]

func (orderBy OrderBy) SQLStr() string {
	return fmt.Sprintf("%s %s", orderBy.A.SQLStr(), orderBy.B)
}

var (
	_ SQL = (*Attr)(nil)
	_ SQL = (*Criteria)(nil)
	_ SQL = (*OrderBy)(nil)
)
