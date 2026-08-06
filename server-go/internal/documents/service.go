package documents

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
	"skillpass-server-go/internal/storage"
)

// MaxDocumentBytes caps a single upload (25 MB).
const MaxDocumentBytes = 25 << 20

var (
	ErrNotFound   = errors.New("document not found")
	ErrTooLarge   = errors.New("file exceeds the 25MB limit")
	ErrBadRequest = errors.New("invalid request")
)

var validCategory = map[string]bool{
	"identity": true, "contract": true, "certificate": true,
	"payslip": true, "tax": true, "other": true,
}

var extPattern = regexp.MustCompile(`^[a-zA-Z0-9.]+$`)

// Anchorer appends a signed, hash-chained integrity record for an entity
// (implemented by the identity service). Wired in main.go; optional.
type Anchorer interface {
	Anchor(ctx context.Context, entityType string, entityID uuid.UUID, sha256Hash string) error
}

// anchorCategories are document categories eligible for integrity anchoring
// after a clean malware scan (Phase 2 · Sprint 6).
var anchorCategories = map[string]bool{
	"contract": true, "certificate": true, "payslip": true,
}

type Service struct {
	db       *sql.DB
	bun      bun.IDB
	store    storage.Store
	scanner  *Scanner
	anchorer Anchorer
}

func NewService(db *sql.DB, bunDB bun.IDB, store storage.Store, scanner *Scanner) *Service {
	return &Service{db: db, bun: bunDB, store: store, scanner: scanner}
}

// SetAnchorer enables integrity anchoring of eligible documents after a clean scan.
func (s *Service) SetAnchorer(a Anchorer) { s.anchorer = a }

// DocumentResponse is the API shape (camelCase, PII-safe).
type DocumentResponse struct {
	ID             string  `json:"id"`
	EmployeeID     *string `json:"employeeId,omitempty"`
	EmployeeName   string  `json:"employeeName,omitempty"`
	Category       string  `json:"category"`
	Filename       string  `json:"filename"`
	MimeType       string  `json:"mimeType"`
	FileSize       int64   `json:"fileSize"`
	ScanStatus     string  `json:"scanStatus"`
	UploadedByName string  `json:"uploadedByName,omitempty"`
	CreatedAt      string  `json:"createdAt"`
} //@name DocumentResponse

// AccessLogResponse is an audit-log row.
type AccessLogResponse struct {
	ID           string `json:"id"`
	DocumentID   string `json:"documentId"`
	Filename     string `json:"filename"`
	AccessedName string `json:"accessedByName,omitempty"`
	Action       string `json:"action"`
	IPAddress    string `json:"ipAddress,omitempty"`
	CreatedAt    string `json:"createdAt"`
} //@name DocumentAccessLog

type UploadParams struct {
	CompanyID  uuid.UUID
	UploadedBy uuid.UUID
	EmployeeID *uuid.UUID
	Category   string
	Filename   string
	MimeType   string
	IPAddress  string
}

// Upload stores a document, records it (scan pending), logs the access, and
// kicks off an async malware scan that flips scan_status.
func (s *Service) Upload(ctx context.Context, p UploadParams, r io.Reader) (*DocumentResponse, error) {
	if !validCategory[p.Category] {
		p.Category = "other"
	}

	// Read (capped) so we can hash, store, and scan the same bytes.
	limited := io.LimitReader(r, MaxDocumentBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	if len(data) == 0 {
		return nil, ErrBadRequest
	}
	if len(data) > MaxDocumentBytes {
		return nil, ErrTooLarge
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])

	docID := uuid.New()
	ext := strings.ToLower(path.Ext(p.Filename))
	if ext != "" && !extPattern.MatchString(strings.TrimPrefix(ext, ".")) {
		ext = "" // drop suspicious extensions
	}
	key := fmt.Sprintf("documents/%s/%s%s", p.CompanyID, docID, ext)

	if _, err := s.store.Save(ctx, key, bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("store document: %w", err)
	}

	doc := &models.Document{
		ID:               docID,
		CompanyID:        p.CompanyID,
		EmployeeID:       p.EmployeeID,
		Category:         p.Category,
		OriginalFilename: p.Filename,
		StorageKey:       key,
		SHA256Hash:       hash,
		MimeType:         p.MimeType,
		FileSize:         int64(len(data)),
		ScanStatus:       "pending",
		UploadedBy:       &p.UploadedBy,
		CreatedAt:        time.Now(),
	}
	if _, err := s.bun.NewInsert().Model(doc).Exec(ctx); err != nil {
		_ = s.store.Delete(ctx, key) // best-effort cleanup
		return nil, fmt.Errorf("insert document: %w", err)
	}

	s.logAccess(ctx, docID, &p.UploadedBy, "upload", p.IPAddress)

	// Async scan (Asynq-ready — swap this goroutine for an enqueue later).
	go s.scanAndUpdate(docID, data)

	return &DocumentResponse{
		ID:         docID.String(),
		Category:   doc.Category,
		Filename:   doc.OriginalFilename,
		MimeType:   doc.MimeType,
		FileSize:   doc.FileSize,
		ScanStatus: doc.ScanStatus,
		CreatedAt:  doc.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) scanAndUpdate(docID uuid.UUID, data []byte) {
	ctx := context.Background()
	status, err := s.scanner.Scan(ctx, bytes.NewReader(data))
	if err != nil {
		slog.Warn("document scan failed", "documentID", docID, "error", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE documents SET scan_status = $1 WHERE id = $2`, status, docID); err != nil {
		slog.Error("update scan status", "documentID", docID, "error", err)
		return
	}
	if status == "clean" {
		s.anchorIfEligible(ctx, docID)
	}
}

// anchorIfEligible appends a signed integrity anchor for eligible document
// categories (contract/certificate/payslip) once they pass the malware scan.
// Best-effort — a signed, tamper-evident record of the file's SHA-256.
func (s *Service) anchorIfEligible(ctx context.Context, docID uuid.UUID) {
	if s.anchorer == nil {
		return
	}
	var category, hash string
	if err := s.db.QueryRowContext(ctx,
		`SELECT category::text, sha256_hash FROM documents WHERE id=$1`, docID).Scan(&category, &hash); err != nil {
		slog.Warn("anchor lookup", "documentID", docID, "error", err)
		return
	}
	if !anchorCategories[category] {
		return
	}
	if err := s.anchorer.Anchor(ctx, "document", docID, hash); err != nil {
		slog.Warn("anchor document", "documentID", docID, "error", err)
	}
}

func (s *Service) logAccess(ctx context.Context, docID uuid.UUID, userID *uuid.UUID, action, ip string) {
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO document_access_log (document_id, accessed_by, action, ip_address) VALUES ($1, $2, $3, $4)`,
		docID, userID, action, ipPtr); err != nil {
		slog.Warn("document access log", "documentID", docID, "action", action, "error", err)
	}
}

// List returns documents for a company, newest first, with optional filters.
func (s *Service) List(ctx context.Context, companyID uuid.UUID, employeeID, category string) ([]DocumentResponse, error) {
	q := `
		SELECT d.id, d.employee_id, COALESCE(e.first_name || ' ' || e.last_name, ''),
		       d.category, d.original_filename, d.mime_type, d.file_size, d.scan_status,
		       COALESCE(u.name, ''), d.created_at
		FROM documents d
		LEFT JOIN employees e ON e.id = d.employee_id
		LEFT JOIN users u ON u.id = d.uploaded_by
		WHERE d.company_id = $1`
	args := []any{companyID}
	if employeeID != "" {
		args = append(args, employeeID)
		q += fmt.Sprintf(" AND d.employee_id = $%d", len(args))
	}
	if category != "" {
		args = append(args, category)
		q += fmt.Sprintf(" AND d.category = $%d", len(args))
	}
	q += " ORDER BY d.created_at DESC LIMIT 500"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []DocumentResponse{}
	for rows.Next() {
		var d DocumentResponse
		var empID *string
		var created time.Time
		if err := rows.Scan(&d.ID, &empID, &d.EmployeeName, &d.Category, &d.Filename,
			&d.MimeType, &d.FileSize, &d.ScanStatus, &d.UploadedByName, &created); err != nil {
			return nil, err
		}
		d.EmployeeID = empID
		d.CreatedAt = created.Format(time.RFC3339)
		out = append(out, d)
	}
	return out, rows.Err()
}

// DownloadTarget carries what the handler needs to stream a document.
type DownloadTarget struct {
	Filename string
	MimeType string
	Body     io.ReadCloser
}

// OpenForDownload verifies ownership, logs the access, and opens the content.
// Infected documents are never served.
func (s *Service) OpenForDownload(ctx context.Context, companyID, docID uuid.UUID, userID *uuid.UUID, ip string) (*DownloadTarget, error) {
	var key, filename, mime, scan string
	err := s.db.QueryRowContext(ctx,
		`SELECT storage_key, original_filename, mime_type, scan_status
		 FROM documents WHERE id = $1 AND company_id = $2`, docID, companyID).
		Scan(&key, &filename, &mime, &scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if scan == "infected" {
		return nil, fmt.Errorf("document failed malware scan")
	}
	body, err := s.store.Open(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("open document: %w", err)
	}
	s.logAccess(ctx, docID, userID, "download", ip)
	return &DownloadTarget{Filename: filename, MimeType: mime, Body: body}, nil
}

// Delete removes the object and the record (company-scoped).
func (s *Service) Delete(ctx context.Context, companyID, docID uuid.UUID) error {
	var key string
	err := s.db.QueryRowContext(ctx,
		`SELECT storage_key FROM documents WHERE id = $1 AND company_id = $2`, docID, companyID).Scan(&key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if err := s.store.Delete(ctx, key); err != nil {
		slog.Warn("delete stored document", "documentID", docID, "error", err)
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = $1 AND company_id = $2`, docID, companyID)
	return err
}

// AuditLog returns recent access-log entries for the company.
func (s *Service) AuditLog(ctx context.Context, companyID uuid.UUID) ([]AccessLogResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.id, l.document_id, d.original_filename, COALESCE(u.name, ''),
		       l.action, COALESCE(l.ip_address, ''), l.created_at
		FROM document_access_log l
		JOIN documents d ON d.id = l.document_id
		LEFT JOIN users u ON u.id = l.accessed_by
		WHERE d.company_id = $1
		ORDER BY l.created_at DESC LIMIT 200`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []AccessLogResponse{}
	for rows.Next() {
		var a AccessLogResponse
		var created time.Time
		if err := rows.Scan(&a.ID, &a.DocumentID, &a.Filename, &a.AccessedName, &a.Action, &a.IPAddress, &created); err != nil {
			return nil, err
		}
		a.CreatedAt = created.Format(time.RFC3339)
		out = append(out, a)
	}
	return out, rows.Err()
}
