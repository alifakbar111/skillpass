package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type IndustryCategory struct {
	bun.BaseModel `bun:"table:industry_categories"`
	ID            uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name          string    `bun:",notnull,unique"`
	Description   *string   `bun:",nullzero"`
}
