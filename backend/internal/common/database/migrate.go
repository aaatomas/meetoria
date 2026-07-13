package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

var skipMigrations = map[string]bool{
	"000_create_databases.sql": true,
}

func Migrate(db *gorm.DB, dir string) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	if err := bootstrapExisting(db); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		if skipMigrations[entry.Name()] {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		applied, err := isMigrationApplied(db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %q: %w", name, err)
		}

		if err := db.Exec(string(body)).Error; err != nil {
			return fmt.Errorf("apply migration %q: %w", name, err)
		}

		if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name).Error; err != nil {
			return fmt.Errorf("record migration %q: %w", name, err)
		}
	}

	return nil
}

func bootstrapExisting(db *gorm.DB) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM schema_migrations").Scan(&count).Error; err != nil {
		return fmt.Errorf("count schema_migrations: %w", err)
	}
	if count > 0 {
		return nil
	}

	var exists bool
	if err := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'organizations'
		)`).Scan(&exists).Error; err != nil {
		return fmt.Errorf("check organizations table: %w", err)
	}
	if !exists {
		return nil
	}

	for _, version := range []string{
		"000_extensions.sql",
		"001_initial_schema.sql",
		"002_service_color.sql",
	} {
		if err := db.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?) ON CONFLICT DO NOTHING",
			version,
		).Error; err != nil {
			return fmt.Errorf("bootstrap migration %q: %w", version, err)
		}
	}

	return nil
}

func isMigrationApplied(db *gorm.DB, version string) (bool, error) {
	var count int64
	if err := db.Raw(
		"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
		version,
	).Scan(&count).Error; err != nil {
		return false, fmt.Errorf("check migration %q: %w", version, err)
	}
	return count > 0, nil
}
