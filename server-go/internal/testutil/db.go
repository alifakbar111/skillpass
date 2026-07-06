package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var TestDBURL = func() string {
	if v := os.Getenv("SKILLPASS_TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/skillpass_test"
}()

var (
	setupOnce sync.Once
	setupErr  error
	globalDB  *sql.DB
)

func SetupTestDB() *sql.DB {
	setupOnce.Do(func() {
		ctx := context.Background()

		// Ensure the test database exists (auto-create if it doesn't)
		if err := ensureTestDBExists(ctx); err != nil {
			setupErr = fmt.Errorf("ensure test db: %w", err)
			return
		}

		db, err := sql.Open("pgx", TestDBURL)
		if err != nil {
			setupErr = fmt.Errorf("open test db: %w", err)
			return
		}
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			setupErr = fmt.Errorf("ping test db: %w", err)
			return
		}
		globalDB = db
		log.Printf("Test DB connected: %s", TestDBURL)
		if err := runMigrations(ctx, db); err != nil {
			db.Close()
			setupErr = fmt.Errorf("run migrations: %w", err)
			return
		}
		log.Println("Test DB migrations complete")
	})
	if setupErr != nil {
		log.Fatalf("Test DB setup failed: %v", setupErr)
	}
	// Clean all tables before each test function for isolation
	CleanTestData(globalDB)
	// Register custom validators needed by handlers
	RegisterTestValidators()
	return globalDB
}

// ensureTestDBExists creates the test database if it doesn't exist.
// Connects to the default "postgres" database to run CREATE DATABASE.
func ensureTestDBExists(ctx context.Context) error {
	u, err := url.Parse(TestDBURL)
	if err != nil {
		return fmt.Errorf("parse test db url: %w", err)
	}
	// Extract the database name (last path segment)
	dbName := u.Path
	if len(dbName) > 0 && dbName[0] == '/' {
		dbName = dbName[1:]
	}
	if dbName == "" {
		return nil // no database name in URL, skip
	}

	// Connect to the default 'postgres' database to create the test DB
	u.Path = "/postgres"
	adminDB, err := sql.Open("pgx", u.String())
	if err != nil {
		return fmt.Errorf("open admin connection: %w", err)
	}
	defer adminDB.Close()

	// Check if the database already exists
	var exists bool
	err = adminDB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check db exists: %w", err)
	}
	if !exists {
		// CREATE DATABASE cannot run inside a transaction
		_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("create database %s: %w", dbName, err)
		}
		log.Printf("Created test database: %s", dbName)
	}
	return nil
}

func NewTestTx(t *testing.T, db *sql.DB) (*sql.Tx, func()) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	return tx, func() { _ = tx.Rollback() }
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	// Resolve the migrations dir from this source file's location so tests in
	// any package find it (a relative "../migrations" silently globs nothing).
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", migrationsDir)
	}
	sort.Strings(files)
	rows, err := db.QueryContext(ctx, "SELECT filename FROM schema_migrations ORDER BY id")
	if err != nil {
		return fmt.Errorf("query applied: %w", err)
	}
	applied := make(map[string]bool)
	for rows.Next() {
		var fn string
		if err := rows.Scan(&fn); err != nil {
			rows.Close()
			return fmt.Errorf("scan: %w", err)
		}
		applied[fn] = true
	}
	rows.Close()
	for _, file := range files {
		filename := filepath.Base(file)
		if applied[filename] {
			continue
		}
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", filename, err)
		}
		if _, err := db.ExecContext(ctx,
			"INSERT INTO schema_migrations (filename) VALUES ($1)", filename,
		); err != nil {
			return fmt.Errorf("record %s: %w", filename, err)
		}
		fmt.Printf("  MIGRATE %s\n", filename)
	}
	return nil
}
