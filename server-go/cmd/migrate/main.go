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

type migration struct {
	file    string
	content string
}

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

	appliedDir := filepath.Join(migrationsDir, ".applied")
	_ = os.MkdirAll(appliedDir, 0755)

	for _, file := range files {
		filename := filepath.Base(file)
		appliedMarker := filepath.Join(appliedDir, filename)

		if _, err := os.Stat(appliedMarker); err == nil {
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

		if err := os.WriteFile(appliedMarker, []byte{}, 0644); err != nil {
			log.Fatalf("Failed to write marker %s: %v", appliedMarker, err)
		}
		fmt.Printf("  OK   %s\n", filename)
	}

	if err := verifySchema(ctx, db); err != nil {
		log.Fatalf("Post-migration verification failed: %v", err)
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("All migrations complete")
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
