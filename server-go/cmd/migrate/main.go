package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type tableDef struct {
	name string
}

var requiredTables = []tableDef{
	{name: "users"},
	{name: "companies"},
	{name: "jobseeker_profiles"},
	{name: "job_experiences"},
	{name: "industry_categories"},
	{name: "tags"},
	{name: "job_postings"},
	{name: "refresh_tokens"},
	{name: "ai_evaluations"},
	{name: "applications"},
}

func main() {
	_ = godotenv.Load(".env", "../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Bootstrap: create schema_migrations table if not exists
	if err := ensureMigrationsTable(ctx, db); err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Transitional import: migrate any existing .applied/ file markers to DB
	if err := importAppliedMarkers(ctx, db); err != nil {
		log.Fatalf("Failed to import .applied/ markers: %v", err)
	}

	migrationsDir := "migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		log.Fatalf("Failed to glob migrations: %v", err)
	}
	if len(files) == 0 {
		log.Println("No migration files found")
		return
	}
	sort.Strings(files)

	// Read already-applied migrations from DB
	applied, err := getAppliedMigrations(ctx, db)
	if err != nil {
		log.Fatalf("Failed to read applied migrations: %v", err)
	}

	for _, file := range files {
		filename := filepath.Base(file)
		if applied[filename] {
			fmt.Printf("  SKIP %s (already applied)\n", filename)
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read %s: %v", filename, err)
		}
		sqlText := string(content)

		if err := runMigration(ctx, db, sqlText); err != nil {
			log.Fatalf("FAILED %s: %v\n\nSQL:\n%s", filename, err, sqlText)
		}

		// Record in DB
		if _, err := db.ExecContext(ctx,
			"INSERT INTO schema_migrations (filename) VALUES ($1)", filename,
		); err != nil {
			log.Fatalf("Failed to record migration %s: %v", filename, err)
		}
		fmt.Printf("  OK   %s\n", filename)
	}

	// Clean up old .applied/ directory
	appliedDir := filepath.Join(migrationsDir, ".applied")
	if _, err := os.Stat(appliedDir); err == nil {
		_ = os.RemoveAll(appliedDir)
	}

	if err := verifySchema(ctx, db); err != nil {
		log.Fatalf("Post-migration verification failed: %v", err)
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("All migrations complete")
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func importAppliedMarkers(ctx context.Context, db *sql.DB) error {
	// Only import if schema_migrations is empty (first run with new system)
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		return fmt.Errorf("count schema_migrations: %w", err)
	}
	if count > 0 {
		return nil // Already seeded
	}

	appliedDir := "migrations/.applied"
	entries, err := os.ReadDir(appliedDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No .applied/ directory, nothing to import
		}
		return fmt.Errorf("read .applied/ directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		modTime := getFileModTime(filepath.Join(appliedDir, entry.Name()))
		_, err := db.ExecContext(ctx,
			"INSERT INTO schema_migrations (filename, applied_at) VALUES ($1, $2)",
			entry.Name(), modTime,
		)
		if err != nil {
			return fmt.Errorf("import marker %s: %w", entry.Name(), err)
		}
		fmt.Printf("  IMPORT %s (from .applied/)\n", entry.Name())
	}
	return nil
}

func getAppliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT filename FROM schema_migrations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}
	return applied, rows.Err()
}

func getFileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now()
	}
	return info.ModTime()
}

func runMigration(ctx context.Context, db *sql.DB, sqlText string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}

	verifyCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var ok bool
	if err = tx.QueryRowContext(verifyCtx, "SELECT 1").Scan(&ok); err != nil {
		return fmt.Errorf("transaction integrity check failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	committed = true
	return nil
}

func verifySchema(ctx context.Context, db *sql.DB) error {
	for _, t := range requiredTables {
		var exists bool
		err := db.QueryRowContext(ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)",
			t.name,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check table %s: %w", t.name, err)
		}
		if !exists {
			return fmt.Errorf("required table missing: %s", t.name)
		}
	}
	return nil
}
