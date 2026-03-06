package database

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all pending migrations.
func RunMigrations(dsn, migrationsPath string) error {
	sourceURL := buildMigrationSource(migrationsPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}
	defer closeMigration(m)

	logMigrationVersion("before up", m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrate up: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("No new migrations to apply")
	} else {
		log.Printf("Migrations applied successfully")
	}

	logMigrationVersion("after up", m)
	return nil
}

// RollbackMigration rolls back only the latest migration step.
func RollbackMigration(dsn, migrationsPath string) error {
	sourceURL := buildMigrationSource(migrationsPath)

	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("create migration instance: %w", err)
	}
	defer closeMigration(m)

	logMigrationVersion("before down", m)

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrate down: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("No migration to rollback")
	} else {
		log.Printf("Rollback executed successfully")
	}

	logMigrationVersion("after down", m)
	return nil
}

func buildMigrationSource(migrationsPath string) string {
	if strings.HasPrefix(migrationsPath, "file://") {
		return migrationsPath
	}

	cleaned := strings.ReplaceAll(migrationsPath, "\\", "/")
	cleaned = path.Clean(cleaned)

	return "file://" + cleaned
}

func logMigrationVersion(stage string, m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Printf("Migration version %s: no version applied yet", stage)
			return
		}
		log.Printf("Migration version %s: error reading version: %v", stage, err)
		return
	}

	log.Printf("Migration version %s: version=%d dirty=%t", stage, version, dirty)
}

func closeMigration(m *migrate.Migrate) {
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil {
		log.Printf("Close migration source error: %v", sourceErr)
	}
	if databaseErr != nil {
		log.Printf("Close migration database error: %v", databaseErr)
	}
}

// CreateMigrationFiles creates the next sequential up/down SQL migration files.
func CreateMigrationFiles(migrationsPath, name string) (string, string, error) {
	if strings.TrimSpace(name) == "" {
		return "", "", fmt.Errorf("migration name is required")
	}

	if err := os.MkdirAll(migrationsPath, 0o755); err != nil {
		return "", "", fmt.Errorf("ensure migrations dir: %w", err)
	}

	next, err := nextMigrationSequence(migrationsPath)
	if err != nil {
		return "", "", err
	}

	safeName := sanitizeMigrationName(name)
	base := fmt.Sprintf("%06d_%s", next, safeName)
	upPath := filepath.Join(migrationsPath, base+".up.sql")
	downPath := filepath.Join(migrationsPath, base+".down.sql")

	if err := os.WriteFile(upPath, []byte("-- Write your UP migration SQL here\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("write up migration file: %w", err)
	}
	if err := os.WriteFile(downPath, []byte("-- Write your DOWN migration SQL here\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("write down migration file: %w", err)
	}

	return upPath, downPath, nil
}

func nextMigrationSequence(migrationsPath string) (int, error) {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}

	maxSeq := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) != 2 {
			continue
		}
		seq, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if seq > maxSeq {
			maxSeq = seq
		}
	}

	return maxSeq + 1, nil
}

func sanitizeMigrationName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "-", "_")
	safe := strings.ToLower(strings.TrimSpace(replacer.Replace(name)))
	if safe == "" {
		return "migration"
	}
	return safe
}
