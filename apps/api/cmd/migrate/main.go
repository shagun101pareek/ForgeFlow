package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/shagun101pareek/forgeflow/internal/config"
)

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}
	command := os.Args[1]

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	source, err := migrationsSource()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve migrations: %v\n", err)
		os.Exit(1)
	}

	m, err := migrate.New(source, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create migrator: %v\n", err)
		os.Exit(1)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
		printVersion(m)
	case "version", "status":
		printVersion(m)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrate <up|version|status>")
}

func printVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version: none")
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read version: %v\n", err)
		os.Exit(1)
	}
	if dirty {
		fmt.Printf("version: %d (dirty)\n", version)
		return
	}
	fmt.Printf("version: %d\n", version)
}

func migrationsSource() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("could not locate migrate binary path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	dir := filepath.Join(root, "sql", "migrations")
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("migrations not found at %s", dir)
	}
	return "file://" + filepath.ToSlash(dir), nil
}
