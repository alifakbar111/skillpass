package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Company struct {
	bun.BaseModel      `bun:"table:companies"`
	ID                 uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID             uuid.UUID  `bun:",notnull,unique"`
	CompanyName        string     `bun:",notnull"`
	Website            *string    `bun:",nullzero"`
	Industry           string     `bun:",notnull"`
	Description        *string    `bun:",nullzero"`
	VerificationStatus string     `bun:",notnull"`
	VerificationDocs   *string    `bun:",nullzero"`
	VerifiedAt         *time.Time `bun:",nullzero"`
	CreatedAt          time.Time  `bun:",notnull"`
	BlindMode          bool       `bun:",notnull"`
}
