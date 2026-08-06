package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Document struct {
	bun.BaseModel    `bun:"table:documents"`
	ID               uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	CompanyID        uuid.UUID  `bun:",notnull"`
	EmployeeID       *uuid.UUID `bun:",nullzero"`
	Category         string     `bun:",notnull"`
	OriginalFilename string     `bun:",notnull"`
	StorageKey       string     `bun:",notnull"`
	SHA256Hash       string     `bun:"sha256_hash,notnull"`
	MimeType         string     `bun:",notnull"`
	FileSize         int64      `bun:",notnull"`
	ScanStatus       string     `bun:",notnull"`
	UploadedBy       *uuid.UUID `bun:",nullzero"`
	CreatedAt        time.Time  `bun:",notnull"`
}

type DocumentAccessLog struct {
	bun.BaseModel `bun:"table:document_access_log"`
	ID            uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	DocumentID    uuid.UUID  `bun:",notnull"`
	AccessedBy    *uuid.UUID `bun:",nullzero"`
	Action        string     `bun:",notnull"`
	IPAddress     *string    `bun:",nullzero"`
	CreatedAt     time.Time  `bun:",notnull"`
}

type ExportJob struct {
	bun.BaseModel `bun:"table:export_jobs"`
	ID            uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	CompanyID     uuid.UUID  `bun:",notnull"`
	Type          string     `bun:",notnull"`
	Status        string     `bun:",notnull"`
	FileURL       *string    `bun:",nullzero"`
	RequestedBy   *uuid.UUID `bun:",nullzero"`
	CreatedAt     time.Time  `bun:",notnull"`
	CompletedAt   *time.Time `bun:",nullzero"`
}
