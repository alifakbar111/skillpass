package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDBWithContainer spins up a PostgreSQL container for CI/local testing.
// Falls back to SKILLPASS_TEST_DATABASE_URL if the env var is set.
func SetupTestDBWithContainer(t *testing.T) *sql.DB {
	t.Helper()

	// If a test DB URL is already configured, use it directly
	if url := os.Getenv("SKILLPASS_TEST_DATABASE_URL"); url != "" {
		return SetupTestDB()
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "skillpass_test",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start test container: %v", err)
	}

	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("postgres://postgres:postgres@%s:%s/skillpass_test?sslmode=disable", host, port.Port())

	os.Setenv("SKILLPASS_TEST_DATABASE_URL", dsn)
	return SetupTestDB()
}
