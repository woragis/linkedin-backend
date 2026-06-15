package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const ensureTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

func ResolveDir() string {
	if d := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); d != "" {
		return filepath.Clean(d)
	}
	candidates := []string{
		"migrations",
		filepath.Join("..", "migrations"),
		filepath.Join("..", "..", "migrations"),
	}
	for _, c := range candidates {
		fi, err := os.Stat(c)
		if err == nil && fi.IsDir() {
			return filepath.Clean(c)
		}
	}
	return ""
}

func Up(ctx context.Context, db *sql.DB, dir string) error {
	if dir == "" {
		return fmt.Errorf("migrations directory is empty")
	}
	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("migrations stat %q: %w", dir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("migrations path %q is not a directory", dir)
	}

	if _, err := db.ExecContext(ctx, ensureTableSQL); err != nil {
		return fmt.Errorf("schema_migrations ddl: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(strings.ToLower(n), ".sql") {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var n int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = $1`, name).Scan(&n); err != nil {
			return fmt.Errorf("migration %s: query applied: %w", name, err)
		}
		if n > 0 {
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("migration %s: read: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("migration %s: begin: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: exec: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: record: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %s: commit: %w", name, err)
		}
	}
	return nil
}
