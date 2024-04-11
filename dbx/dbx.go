// nolint
package dbx

import (
	"context"
	"database/sql"
	"strings"

	"github.com/kcmvp/gob/internal"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
)

const DefaultDS = "defaultDS"

// DBX database adapter
type DBX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Close() error
}

type Hook func(sql string) string

type DBXImpl struct { //nolint
	ds               DBX
	beforeQueryHooks []Hook
	beforeExecHooks  []Hook
}

func (dbxImpl *DBXImpl) PoolSize() int32 {
	// TODO implement me
	panic("implement me")
}

func (dbxImpl *DBXImpl) TotalConns() int32 {
	// TODO implement me
	panic("implement me")
}

func (dbxImpl *DBXImpl) IdleConns() int32 {
	// TODO implement me
	panic("implement me")
}

func (dbxImpl *DBXImpl) MaxIdleDestroyCount() int32 {
	// TODO implement me
	panic("implement me")
}

func (dbxImpl *DBXImpl) Close() error {
	return dbxImpl.ds.Close()
}

func (dbxImpl *DBXImpl) Shutdown() {
	dbxImpl.Close()
}

func (dbxImpl *DBXImpl) HealthCheck(_ context.Context) error {
	panic("print pool status")
}

func (dbxImpl *DBXImpl) PrepareContext(_ context.Context, s string) (*sql.Stmt, error) {
	// TODO implement me
	panic("implement me")
}

func (dbxImpl *DBXImpl) ExecContext(_ context.Context, s string, i ...interface{}) (sql.Result, error) {
	// TODO implement me
	for _, hook := range dbxImpl.beforeExecHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *DBXImpl) QueryContext(ctx context.Context, s string, i ...interface{}) (*sql.Rows, error) {
	for _, hook := range dbxImpl.beforeQueryHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *DBXImpl) QueryRowContext(ctx context.Context, s string, i ...interface{}) *sql.Row {
	for _, hook := range dbxImpl.beforeQueryHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *DBXImpl) AddQueryHook(hook Hook) {
	dbxImpl.beforeQueryHooks = append(dbxImpl.beforeQueryHooks, hook)
}

func (dbxImpl *DBXImpl) AddExecHooks(hook Hook) {
	dbxImpl.beforeExecHooks = append(dbxImpl.beforeExecHooks, hook)
}

func init() {
	lo.ForEach(internal.Container.ListProvidedServices(), func(item do.EdgeService, _ int) {
		if strings.HasSuffix(item.Service, "_*database/sql.DB") {
			dsName := strings.TrimSuffix(item.Service, "_*database/sql.DB")
			do.ProvideNamed[DBX](internal.Container, dsName, func(injector do.Injector) (DBX, error) {
				return &DBXImpl{ds: do.MustInvokeNamed[*sql.DB](injector, item.Service)}, nil
			})
		}
	})
}

var (
	_ DBX                         = (*DBXImpl)(nil)
	_ do.HealthcheckerWithContext = (*DBXImpl)(nil)
	_ do.Shutdowner               = (*DBXImpl)(nil)
)
