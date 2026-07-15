package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Notification struct {
	bun.BaseModel `bun:"table:notifications"`
	ID        uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID  `bun:",notnull"`
	Type      string     `bun:",notnull"`
	Title     string     `bun:",notnull"`
	Body      string     `bun:",notnull"`
	Link      string     `bun:",notnull"`
	ReadAt    *time.Time `bun:",nullzero"`
	CreatedAt time.Time  `bun:",notnull"`
}
