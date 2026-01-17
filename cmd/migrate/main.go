package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	var migrationsPath string
	flag.StringVar(&migrationsPath, "path", "pkg/database/migrations", "path to migrations")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: migrate [up|down|version|force N]")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://sovereign:sovereign123@localhost:5432/sovereign_firm?sslmode=disable"
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	cmd := flag.Args()[0]

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migrations rolled back successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if len(flag.Args()) < 2 {
			log.Fatal("force requires a version number")
		}
		var v int
		fmt.Sscanf(flag.Args()[1], "%d", &v)
		if err := m.Force(v); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		log.Printf("Forced to version %d", v)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
