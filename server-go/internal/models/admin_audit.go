package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type AdminAudit struct {
	bun.BaseModel `bun:"table:admin_audit_log"`
	ID        uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	AdminID   uuid.UUID  `bun:",notnull"`
	CompanyID uuid.UUID  `bun:",notnull"`
	Action    string     `bun:",notnull"`
	Reason    *string    `bun:",nullzero"`
	CreatedAt time.Time  `bun:",notnull"`
}
