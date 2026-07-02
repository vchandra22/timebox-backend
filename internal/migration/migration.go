package migration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

const createSchemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

func Run(ctx context.Context, db *sqlx.DB, dir string) ([]string, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, createSchemaMigrationsTable); err != nil {
		return nil, err
	}

	files, err := migrationFiles(dir)
	if err != nil {
		return nil, err
	}

	applied := make([]string, 0, len(files))
	for _, file := range files {
		version := filepath.Base(file)

		var exists bool
		if err := tx.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version); err != nil {
			return nil, err
		}
		if exists {
			continue
		}

		query, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if err := execStatements(ctx, tx, string(query)); err != nil {
			return nil, fmt.Errorf("apply %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return nil, err
		}
		applied = append(applied, version)
	}

	return applied, tx.Commit()
}

func migrationFiles(dir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func execStatements(ctx context.Context, tx *sqlx.Tx, query string) error {
	// ponytail: simple SQL files only; use a real migration tool if function bodies need semicolon-aware parsing.
	for _, statement := range strings.Split(query, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
