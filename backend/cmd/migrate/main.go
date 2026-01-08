package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

func main() {
	var (
		direction = flag.String("dir", "up", "Migration direction: up or down")
		dbURL     = flag.String("db", "", "Database URL (or set DATABASE_URL env)")
	)
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, *dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Ensure migrations table exists
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try from backend directory
		migrationsDir = "backend/migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	suffix := "." + *direction + ".sql"
	var migrations []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), suffix) {
			migrations = append(migrations, f.Name())
		}
	}
	sort.Strings(migrations)

	if *direction == "down" {
		// Reverse order for down migrations
		for i, j := 0, len(migrations)-1; i < j; i, j = i+1, j-1 {
			migrations[i], migrations[j] = migrations[j], migrations[i]
		}
	}

	for _, m := range migrations {
		version := strings.Split(m, "_")[0]

		var applied bool
		if *direction == "up" {
			err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
			if err != nil {
				log.Fatalf("Failed to check migration status: %v", err)
			}
			if applied {
				fmt.Printf("Skipping %s (already applied)\n", m)
				continue
			}
		} else {
			err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&applied)
			if err != nil {
				log.Fatalf("Failed to check migration status: %v", err)
			}
			if !applied {
				fmt.Printf("Skipping %s (not applied)\n", m)
				continue
			}
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, m))
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", m, err)
		}

		fmt.Printf("Running %s...\n", m)
		_, err = conn.Exec(ctx, string(content))
		if err != nil {
			log.Fatalf("Failed to run migration %s: %v", m, err)
		}

		if *direction == "up" {
			_, err = conn.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
		} else {
			_, err = conn.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version)
		}
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", m, err)
		}

		fmt.Printf("Completed %s\n", m)
	}

	fmt.Println("Migrations complete!")
}
