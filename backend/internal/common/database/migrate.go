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

	if err := linkExistingBranchServices(db); err != nil {
		return fmt.Errorf("link branch services: %w", err)
	}

	return nil
}

func linkExistingBranchServices(db *gorm.DB) error {
	return db.Exec(`
		INSERT INTO branch_services (organization_id, branch_id, service_id)
		SELECT s.organization_id, b.id, s.id
		FROM services s
		JOIN branches b ON b.organization_id = s.organization_id AND b.is_default = true
		ON CONFLICT (branch_id, service_id) DO NOTHING
	`).Error
}

func bootstrapExisting(db *gorm.DB) error {
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

	applied, err := isMigrationApplied(db, "001_schema.sql")
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	if err := db.Exec(
		"INSERT INTO schema_migrations (version) VALUES (?)",
		"001_schema.sql",
	).Error; err != nil {
		return fmt.Errorf("bootstrap migration %q: %w", "001_schema.sql", err)
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
