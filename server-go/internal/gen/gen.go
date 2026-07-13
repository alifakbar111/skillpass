package gen

//go:generate jet -source=postgres -dsn="postgres://postgres:postgres@localhost:5432/skillpass?sslmode=disable" -schema=public -path=../../.gen -ignore-tables=__drizzle_migrations
