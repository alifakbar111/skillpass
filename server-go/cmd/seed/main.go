package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
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
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	fmt.Println("Seeding database...")

	industries := []struct {
		Name        string
		Description string
	}{
		{"Technology", "Software, hardware, IT services"},
		{"Manufacturing", "Industrial production and fabrication"},
		{"Healthcare", "Medical services and pharmaceuticals"},
		{"Finance", "Banking, investment, insurance"},
		{"Education", "Schools, universities, training"},
		{"Retail", "Sales, e-commerce, consumer goods"},
		{"Transportation", "Logistics, delivery, ride-hailing"},
		{"Creative Arts", "Design, media, entertainment"},
		{"Hospitality", "Hotels, restaurants, tourism"},
		{"Construction", "Building and infrastructure"},
		{"Agriculture", "Farming, food production"},
		{"Energy", "Oil, gas, renewable energy"},
	}

	count := 0
	for _, ind := range industries {
		stmt := gen.IndustryCategories.INSERT(
			gen.IndustryCategories.Name, gen.IndustryCategories.Description,
		).VALUES(ind.Name, ind.Description).
			ON_CONFLICT(gen.IndustryCategories.Name).DO_NOTHING()

		_, err := stmt.ExecContext(ctx, db)
		if err != nil {
			log.Printf("  Warning: failed to insert industry %s: %v", ind.Name, err)
			continue
		}
		count++
	}
	fmt.Printf("Seeded %d industry categories\n", count)

	checkStmt := SELECT(
		gen.Users.ID,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.Email.EQ(String("admin-skillpass@yopmail.com")),
	)

	var existing model.Users
	err = checkStmt.QueryContext(ctx, db, &existing)
	if err == nil {
		fmt.Println("Admin user already exists, skipping")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin123!!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	insertStmt := gen.Users.INSERT(
		gen.Users.Email, gen.Users.Username, gen.Users.PasswordHash,
		gen.Users.Name, gen.Users.Role,
	).VALUES(
		"admin-skillpass@yopmail.com", "admin", string(passwordHash), "Admin", "admin",
	)

	_, err = insertStmt.ExecContext(ctx, db)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	fmt.Println("Seeded admin user (admin-skillpass@yopmail.com / admin123!!)")
}
