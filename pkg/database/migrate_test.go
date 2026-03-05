package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMigrationSource(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "relative path", in: "migrations", want: "file://migrations"},
		{name: "windows style path", in: `migrations\\sub`, want: "file://migrations/sub"},
		{name: "absolute unix path", in: "/tmp/migrations", want: "file:///tmp/migrations"},
		{name: "already url", in: "file://migrations", want: "file://migrations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildMigrationSource(tt.in)
			if got != tt.want {
				t.Fatalf("buildMigrationSource(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestInitialMigrationSQLContent(t *testing.T) {
	upPath := filepath.Join("..", "..", "migrations", "000001_init_schema.up.sql")
	downPath := filepath.Join("..", "..", "migrations", "000001_init_schema.down.sql")

	upData, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}
	downData, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	upSQL := strings.ToLower(string(upData))
	downSQL := strings.ToLower(string(downData))

	if !strings.Contains(upSQL, "create extension") {
		t.Fatalf("up migration should contain CREATE EXTENSION statement")
	}
	if !strings.Contains(upSQL, "uuid-ossp") {
		t.Fatalf("up migration should reference uuid-ossp extension")
	}
	if !strings.Contains(downSQL, "drop extension") {
		t.Fatalf("down migration should contain DROP EXTENSION statement")
	}
	if !strings.Contains(downSQL, "uuid-ossp") {
		t.Fatalf("down migration should reference uuid-ossp extension")
	}
}

func TestRunMigrationsAndRollback_Integration(t *testing.T) {
	dsn := os.Getenv("MIGRATION_TEST_DB_URL")
	if dsn == "" {
		t.Skip("MIGRATION_TEST_DB_URL not set; skipping integration migration test")
	}

	path := filepath.Join("..", "..", "migrations")
	if err := RunMigrations(dsn, path); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	if err := RollbackMigration(dsn, path); err != nil {
		t.Fatalf("RollbackMigration failed: %v", err)
	}
}
