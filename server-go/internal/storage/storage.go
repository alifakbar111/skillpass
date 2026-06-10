// Package storage abstracts user-file persistence (avatars, future uploads).
// It mirrors the lib.LLMClient pattern: small interface, env-driven factory.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Store persists a file and returns its public URL path.
type Store interface {
	// Save writes the content under key (e.g. "avatars/<id>.png") and returns
	// the URL path clients can fetch it from (e.g. "/uploads/avatars/<id>.png").
	Save(ctx context.Context, key string, r io.Reader) (string, error)
}

var _ Store = (*LocalStore)(nil)

// NewStore creates a Store based on STORAGE_PROVIDER. Currently only the
// local-disk implementation exists; an S3-compatible store can be added
// behind the same interface without touching callers.
func NewStore() Store {
	// switch os.Getenv("STORAGE_PROVIDER") { case "s3": ... }
	return NewLocalStore()
}

// LocalStore writes files to a directory on disk, served by the API server
// under /uploads/. Directory from UPLOAD_DIR (default ./uploads).
type LocalStore struct {
	dir string
}

func NewLocalStore() *LocalStore {
	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = "./uploads"
	}
	return &LocalStore{dir: dir}
}

// Dir returns the root directory, used by main.go to register the static route.
func (l *LocalStore) Dir() string {
	return l.dir
}

var keyPattern = regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`)

func (l *LocalStore) Save(_ context.Context, key string, r io.Reader) (string, error) {
	// Reject anything that could escape the upload root.
	if !keyPattern.MatchString(key) || strings.Contains(key, "..") || strings.HasPrefix(key, "/") {
		return "", fmt.Errorf("invalid storage key %q", key)
	}

	fullPath := filepath.Join(l.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create upload file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("write upload: %w", err)
	}

	return "/uploads/" + key, nil
}
