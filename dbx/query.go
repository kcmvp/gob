// nolint
package dbx

import (
	"github.com/samber/lo"
)

type Joint lo.Tuple3[Attr, Attr, string]

func (joint Joint) String() string {
	return ""
}

type Result struct {
	raw   map[string]any
	attrs []Attr //nolint
}

func (result Result) Get(attr Attr) (any, error) {
	return nil, nil
}

func (result Result) Count() int {
	return len(result.raw)
}

type Query struct {
	jointStr string
}

func Select(attrs []Attr) Query {
	//@todo validate
	return Query{}
}

func (query Query) WithJoin(joints []Joint) Query {
	//@todo validate
	return Query{}
}

func (query Query) OrderBy(orders []OrderBy) Query {
	return Query{}
}

func (query Query) Where(criteria Criteria, values func() []any) Query {
	return Query{}
}

func (query Query) Rows() Result {
	return Result{}
}
