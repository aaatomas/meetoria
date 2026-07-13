package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, dir string) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
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
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		var count int64
		if err := db.Raw(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			name,
		).Scan(&count).Error; err != nil {
			return fmt.Errorf("check migration %q: %w", name, err)
		}
		if count > 0 {
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
