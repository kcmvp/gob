package repository

import (
	"fmt"
	"github.com/kcmvp/dbo"
	. "github.com/kcmvp/dbo/entity"
	"github.com/kcmvp/dbo/internal"
	"github.com/samber/do/v2"
)

type Repository[E IEntity] interface {
	Insert(entity E) (int64, error)
	BatchInsert(entities []E) (int64, error)
	Delete(key any) (int64, error)
	Update(set []Set[E], where Where[E]) (int64, error)
	Find(key any) (E, error)
	FindBy(where Where[E]) (E, error)
	DeleteBy(where Where[E]) (int64, error)
	Select(columns []Column[E], where Where[E], orders ...OrderBy) ([]E, error)
	Join(joins ...Join[IEntity, IEntity]) SqlBuilder
}

type MustBeStructError struct {
	msg string
}

func (e MustBeStructError) Error() string {
	return fmt.Sprintf("must be struct %s", e.msg)
}

// defaultRepository default Repository implementation
type defaultRepository[E IEntity] struct {
	zeroE E
	dbo.DBO
}

func (d defaultRepository[E]) BatchInsert(entities []E) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Delete(key any) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Update(set []Set[E], where Where[E]) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Find(key any) (E, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) FindBy(where Where[E]) (E, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) DeleteBy(where Where[E]) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Select(columns []Column[E], where Where[E], orders ...OrderBy) ([]E, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Insert(entity E) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (d defaultRepository[E]) Join(joins ...Join[IEntity, IEntity]) SqlBuilder {
	panic("implement me")
}

func NewRepositoryWithDS[E IEntity](dsName string) do.Provider[Repository[E]] {
	return func(inject do.Injector) (Repository[E], error) {
		return &defaultRepository[E]{
			DBO: do.MustInvokeNamed[dbo.DBO](inject, dsName),
		}, nil
	}
}

func NewRepository[E IEntity]() do.Provider[Repository[E]] {
	return NewRepositoryWithDS[E](internal.DefaultDS)
}
