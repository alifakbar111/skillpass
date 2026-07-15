package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type User struct {
	bun.BaseModel `bun:"table:users"`
	ID           uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Email        string    `bun:",notnull,unique"`
	Username     string    `bun:",notnull,unique"`
	PasswordHash string    `bun:",notnull"`
	Role         string    `bun:",notnull"`
	Name         string    `bun:",notnull"`
	AvatarURL    *string   `bun:",nullzero"`
	IsVerified   bool      `bun:",notnull"`
	CreatedAt    time.Time `bun:",notnull"`
}
