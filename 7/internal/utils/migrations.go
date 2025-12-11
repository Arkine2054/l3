package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunMigrations(db *sql.DB, migrateDir string) error {
	files, err := os.ReadDir(migrateDir)
	if err != nil {
		return fmt.Errorf("read migrate dir: %w", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}

		path := filepath.Join(migrateDir, f.Name())
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", path, err)
		}

		fmt.Println("Applied migration:", f.Name())
	}
	return nil
}
