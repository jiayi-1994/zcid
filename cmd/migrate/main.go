package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xjy/zcid/config"
	"github.com/xjy/zcid/pkg/database"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run cmd/migrate/main.go [up|down|new] [--name migration_name]")
	}

	command := os.Args[1]
	switch command {
	case "up":
		runUp()
	case "down":
		runDown()
	case "new":
		runNew(os.Args[2:])
	default:
		log.Fatalf("Unknown migrate command: %s", command)
	}
}

func runUp() {
	dbURL, err := resolveDatabaseURL()
	if err != nil {
		log.Fatalf("Resolve database URL failed: %v", err)
	}

	if err := database.RunMigrations(dbURL, "migrations"); err != nil {
		log.Fatalf("Run migrations failed: %v", err)
	}

	log.Println("Migrate up completed")
}

func runDown() {
	dbURL, err := resolveDatabaseURL()
	if err != nil {
		log.Fatalf("Resolve database URL failed: %v", err)
	}

	if err := database.RollbackMigration(dbURL, "migrations"); err != nil {
		log.Fatalf("Rollback migration failed: %v", err)
	}

	log.Println("Migrate down completed")
}

func runNew(args []string) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	name := fs.String("name", "", "migration name")
	_ = fs.Parse(args)

	if *name == "" {
		log.Fatal("Missing --name for migrate new")
	}

	upPath, downPath, err := database.CreateMigrationFiles("migrations", *name)
	if err != nil {
		log.Fatalf("Create migration files failed: %v", err)
	}

	fmt.Printf("Created migration files:\n- %s\n- %s\n", upPath, downPath)
}

func resolveDatabaseURL() (string, error) {
	if dbURL := os.Getenv("DB_URL"); dbURL != "" {
		return dbURL, nil
	}

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		return "", fmt.Errorf("load config: %w", err)
	}

	return cfg.Database.MigrationURL(), nil
}
