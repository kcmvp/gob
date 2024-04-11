package dbx

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type Base struct {
	CreatedAt time.Time `ds:"autoUpdateTime"`
	CreatedBy string    `ds:"createdBy"`
	UpdatedAt time.Time `ds:"autoCreateTime"`
	UpdatedBy string    `ds:"updatedBy"`
}

type Product struct {
	Base
	Id          string `ds:"pk;"`
	Name        string `ds:"column=name"`
	FullName    string `ds:"ignore"`
	Grade       sql.NullInt32
	Address     sql.NullString
	ProductDate time.Time
	comment     string
}

func (p Product) Table() string {
	return "product"
}

func TestName(t *testing.T) {
	//assert.Equal(t, 1, 1)
}

type BuilderTestSuit struct {
	//builder *SqlBuilder
	suite.Suite
}

func TestBuilderSuite(t *testing.T) {
	suite.Run(t, &BuilderTestSuit{})
}

//	func (suite *BuilderTestSuit) SetupSuite() {
//		//suite.builder = do.MustInvoke[*SqlBuilder](boot.Container())
//	}
func (suite *BuilderTestSuit) TestHappyFlow() {
	assert.Equal(suite.T(), 1, 1)
}
