// nolint
package dbx

import (
	"fmt"

	"github.com/kcmvp/gob/internal"
	"github.com/samber/do/v2"
)

type Repository[E IEntity, K Key] interface {
	Insert(entity E) (int64, error)
	BatchInsert(entities []E) (int64, error)
	Delete(Key K) (int64, error)
	Update(set []Set, criteria Criteria, values func() []any) (int64, error)
	Find(Key K) (E, error)
	FindBy(criteria Criteria, values func() []any) (E, error)
	DeleteBy(criteria Criteria, values func() []any) (int64, error)
	SearchBy(criteria Criteria, values func() []any, orderBy ...OrderBy) ([]E, error)
}

type MustBeStructError struct {
	msg string
}

func (e MustBeStructError) Error() string {
	return fmt.Sprintf("must be struct %s", e.msg)
}

// defaultRepository default Repository implementation
type defaultRepository[E IEntity, K Key] struct {
	zeroK K //nolint
	zeroE E //nolint
	// sqlBuilder *SqlBuilder
	dbx DBX
}

func (d defaultRepository[E, K]) Insert(entity E) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) BatchInsert(entities []E) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) Delete(Key K) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) Find(Key K) (E, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) FindBy(criteria Criteria, parameters func() []any) (E, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) DeleteBy(criteria Criteria, parameters func() []any) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) Update(set []Set, criteria Criteria, parameters func() []any) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (d defaultRepository[E, K]) SearchBy(criteria Criteria, parameters func() []any, orderBy ...OrderBy) ([]E, error) {
	// TODO implement me
	panic("implement me")
}

func NewRepository[E IEntity, K Key]() Repository[E, K] {
	return NewRepositoryWithDS[E, K](DefaultDS)
}

func NewRepositoryWithDS[E IEntity, K Key](dsName string) Repository[E, K] {
	repo := &defaultRepository[E, K]{
		dbx: do.MustInvokeNamed[DBX](internal.Container, dsName),
	}
	//@todo validate with Attr
	return repo
}
