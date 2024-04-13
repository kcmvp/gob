package internal

import (
	"database/sql"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type Base struct {
	CreatedAt time.Time `db:"act;name=ct_time"`
	CreatedBy string    `db:"type=varchar(20)"`
	UpdatedAt time.Time `db:"aut"`
	UpdatedBy string    `db:"type=varchar(20)"`
}
type Base1 struct {
	Id        int64 `db:"pk;type=integer"`
	CreatedAt time.Time
	UpdatedAt time.Time `db:"ignore"`
}
type Base2 struct {
	CreatedAt time.Time
	UpdatedAt time.Time `db:"ignore;name=abc;type=varchar(20)"`
}

type Product struct {
	Base1
	Address sql.NullString `db:"name=full_address"`
	comment string
}

func TestParse(t *testing.T) {

	tests := []struct {
		name string
		arg  any
		want []string
	}{
		{
			name: "Simple",
			arg:  Base{},
			want: []string{
				Mapper{"CreatedAt", "ct_time", []string{"act", "name=ct_time"}}.String(),
				Mapper{"CreatedBy", "created_by", []string{"type=varchar(20)"}}.String(),
				Mapper{"UpdatedAt", "updated_at", []string{"aut"}}.String(),
				Mapper{"UpdatedBy", "updated_by", []string{"type=varchar(20)"}}.String(),
			},
		},
		{
			name: "Ignore_case1",
			arg:  Base1{},
			want: []string{
				Mapper{A: "CreatedAt", B: "created_at"}.String(),
				Mapper{A: "Id", B: "id", C: []string{"pk", "type=integer"}}.String(),
			},
		},
		{
			name: "Ignore_case2",
			arg:  Base2{},
			want: []string{
				Mapper{A: "CreatedAt", B: "created_at"}.String(),
			},
		},
		{
			name: "embedded",
			arg:  Product{},
			want: []string{
				Mapper{A: "CreatedAt", B: "created_at"}.String(),
				Mapper{A: "Id", B: "id", C: []string{"pk", "type=integer"}}.String(),
				Mapper{A: "Address", B: "full_address", C: []string{"name=full_address"}}.String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lo.Map(Parse(tt.arg), func(item Mapper, _ int) string {
				return item.String()
			})
			assert.True(t, lo.Every(got, tt.want))
		})
	}
}
