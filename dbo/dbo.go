package dbo

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/kcmvp/dbo/internal"
	"github.com/kcmvp/gob/runtime"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"os"
	"path/filepath"
)

// DBO database adapter
type DBO interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	Close() error
}

type Hook func(sql string) string

type dboImpl struct {
	dbo              DBO
	beforeQueryHooks []Hook
	beforeExecHooks  []Hook
}

func (dbxImpl *dboImpl) PoolSize() int32 {
	//TODO implement me
	panic("implement me")
}

func (dbxImpl *dboImpl) TotalConns() int32 {
	//TODO implement me
	panic("implement me")
}

func (dbxImpl *dboImpl) IdleConns() int32 {
	//TODO implement me
	panic("implement me")
}

func (dbxImpl *dboImpl) MaxIdleDestroyCount() int32 {
	//TODO implement me
	panic("implement me")
}

func (dbxImpl *dboImpl) Close() error {
	return dbxImpl.dbo.Close()
}

func (dbxImpl *dboImpl) Shutdown() {
	dbxImpl.Close()
}

func (dbxImpl *dboImpl) HealthCheck(ctx context.Context) error {
	panic("print pool status")
}

func (dbxImpl *dboImpl) PrepareContext(ctx context.Context, s string) (*sql.Stmt, error) {
	//TODO implement me
	panic("implement me")
}

func (dbxImpl *dboImpl) ExecContext(ctx context.Context, s string, i ...interface{}) (sql.Result, error) {
	//TODO implement me
	for _, hook := range dbxImpl.beforeExecHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *dboImpl) QueryContext(ctx context.Context, s string, i ...interface{}) (*sql.Rows, error) {
	for _, hook := range dbxImpl.beforeQueryHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *dboImpl) QueryRowContext(ctx context.Context, s string, i ...interface{}) *sql.Row {
	for _, hook := range dbxImpl.beforeQueryHooks {
		s = hook(s)
	}
	panic("implement me")
}

func (dbxImpl *dboImpl) AddQueryHook(hook Hook) {
	dbxImpl.beforeQueryHooks = append(dbxImpl.beforeQueryHooks, hook)
}
func (dbxImpl *dboImpl) AddExecHooks(hook Hook) {
	dbxImpl.beforeExecHooks = append(dbxImpl.beforeExecHooks, hook)
}

// init initialize dbo objects
func init() {
	for name, ds := range internal.DSMap() {
		if db, err := sql.Open(ds.Driver, ds.DSN()); err == nil {
			if err = db.Ping(); err != nil {
				_ = db.Close()
				panic(fmt.Errorf("failed to initialize %s: %w", name, err))
			}
			// init database with scripts
			if ds.InitDB {
				lo.ForEach(ds.Scripts, func(script string, _ int) {
					if data, err := os.ReadFile(filepath.Join(context.RootDir(), script)); err == nil {
						if _, err = db.Exec(string(data)); err != nil {
							panic(fmt.Errorf("failed to execute %s: %w", script, err))
						}
					} else {
						panic(fmt.Errorf("failed to read %s: %w", script, err))
					}
				})
			}
			injector := do.LazyNamed(ds.DB, func(injector do.Injector) (DBO, error) {
				return &dboImpl{dbo: db}, nil
			})
			context.Register(injector)
		}
	}
}

var _ DBO = (*dboImpl)(nil)
var _ do.HealthcheckerWithContext = (*dboImpl)(nil)
var _ do.Shutdowner = (*dboImpl)(nil)
