package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Tag struct {
	bun.BaseModel      `bun:"table:tags"`
	ID                 uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name               string     `bun:",notnull"`
	IndustryCategoryID *uuid.UUID `bun:",nullzero"`
}
