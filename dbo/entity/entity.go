package entity

import (
	"time"
)

type IEntity interface {
	Table() string
}

type Key interface {
	int64 | string
}

type Entity[K Key] struct {
	Id        K         `json:"id" db:"col=id;pk;seq=-1"`
	UpdatedBy string    `json:"updatedBy" db:"col=updated_by;seq=1000"`
	UpdatedAt time.Time `json:"updatedAt" db:"col=updated_at;aut;seq=1001"`
	CreatedBy string    `json:"createdBy" db:"col=created_by;seq=1002"`
	CreatedAt time.Time `json:"createdAt" db:"col=created_at;act;seq=1003"`
}
