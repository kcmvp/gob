package meta

import (
	"fmt"
	"github.com/kcmvp/app"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestLoadEntity(t *testing.T) {
	result := Build()
	er := result.MustGet()
	assert.Len(t, er.Table("Product").columns, 8)
	assert.Len(t, er.Table("Order").columns, 8)
	assert.Len(t, er.Table("OrderItem").columns, 10)
	assert.Len(t, er.Table("Customer").columns, 9)
	assert.Len(t, er.Table("Address").columns, 8)
	if err := er.ER(filepath.Join(app.RootDir(), "target")); err != nil {
		fmt.Println(err)
	}
	if err := er.Schema(filepath.Join(app.RootDir(), "target")); err != nil {
		fmt.Println(err)
	}
}

func TestDBO_Columns(t *testing.T) {
	result := Build()
	er := result.MustGet()
	er.Columns(filepath.Join(app.RootDir(), "target"))
}
