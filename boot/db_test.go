package boot

import (
	"database/sql"
	"fmt"
	"github.com/kcmvp/gob/internal"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/do/v2"
	typetostring "github.com/samber/go-type-to-string"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DBTestSuite struct {
	suite.Suite
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, &DBTestSuite{})
}

func (dbs *DBTestSuite) SetupSuite() {
	InitApp()
}

func (dbs *DBTestSuite) TestMultipleDB() {
	ds := dsMap()
	assert.Equal(dbs.T(), 2, len(ds))
	db := do.MustInvokeNamed[*sql.DB](internal.Container, fmt.Sprintf("%s_%s", "ds1", typetostring.GetType[*sql.DB]()))
	assert.NotNil(dbs.T(), db)
	rs, err := db.Exec("select * from Product")
	assert.NoError(dbs.T(), err)
	cnt, _ := rs.RowsAffected()
	assert.Equal(dbs.T(), int64(2), cnt)
	db = do.MustInvokeNamed[*sql.DB](internal.Container, fmt.Sprintf("%s_%s", "ds2", typetostring.GetType[*sql.DB]()))
	assert.NotNil(dbs.T(), db)
	rs, err = db.Exec("select * from Product")
	assert.NoError(dbs.T(), err)
	cnt, _ = rs.RowsAffected()
	assert.Equal(dbs.T(), int64(0), cnt)
}
