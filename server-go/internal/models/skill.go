package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Skill struct {
	bun.BaseModel `bun:"table:skills"`
	ID        uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name      string    `bun:",notnull,unique"`
	CreatedAt time.Time `bun:",notnull"`
}
