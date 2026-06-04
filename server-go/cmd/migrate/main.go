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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/skillpass"
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
	os.MkdirAll(appliedDir, 0755)

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

		sql := string(content)

		_, err = db.ExecContext(ctx, sql)
		if err != nil {
			log.Fatalf("FAILED %s: %v\n\nSQL:\n%s", filename, err, sql)
		}

		if err := os.WriteFile(appliedMarker, []byte{}, 0644); err != nil {
			log.Fatalf("Failed to write marker %s: %v", appliedMarker, err)
		}

		fmt.Printf("  OK   %s\n", filename)
	}

	if _, err := db.ExecContext(ctx, "SELECT 1"); err != nil {
		log.Fatalf("Post-migration verification failed: %v", err)
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("All migrations complete")
}
