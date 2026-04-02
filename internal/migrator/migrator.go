package migrator

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Run applies all *.sql files in dir in lexicographic order, skipping versions
// already present in schema_migrations. The table must be created by
// 0000_schema_migrations.sql (or equivalent) on first run.
func Run(db *sql.DB, driver string, dir string) error {
	if err := dirOK(dir); err != nil {
		return err
	}

	applied, err := loadApplied(db, driver)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .sql files in %s", dir)
	}
	sort.Strings(files)

	insertSQL := insertVersionSQL(driver)
	appliedCount := 0

	for _, path := range files {
		name := filepath.Base(path)
		if _, ok := applied[name]; ok {
			continue
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}

		stmts := splitSQLStatements(string(body))
		if len(stmts) == 0 {
			return fmt.Errorf("%s: no executable statements", name)
		}

		if err := runInTx(db, driver, stmts, name, insertSQL); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}

		appliedCount++
		fmt.Printf("applied %s\n", name)
	}

	if appliedCount == 0 {
		fmt.Println("no pending migrations")
	} else {
		fmt.Printf("migrations finished (%d applied)\n", appliedCount)
	}

	return nil
}

func dirOK(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("migrations dir %q: %w", dir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("migrations path %q is not a directory", dir)
	}
	return nil
}

func loadApplied(db *sql.DB, driver string) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		if isNoMigrationsTable(err, driver) {
			return out, nil
		}
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = struct{}{}
	}
	return out, rows.Err()
}

func isNoMigrationsTable(err error, driver string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch driver {
	case "sqlite":
		return strings.Contains(msg, "no such table") && strings.Contains(msg, "schema_migrations")
	case "postgres":
		return strings.Contains(msg, "schema_migrations") && strings.Contains(msg, "does not exist")
	case "mysql":
		// Error 1146: Table 'db.schema_migrations' doesn't exist
		return strings.Contains(msg, "1146") || strings.Contains(msg, "doesn't exist")
	default:
		return strings.Contains(msg, "no such table") ||
			strings.Contains(msg, "does not exist") ||
			strings.Contains(msg, "doesn't exist")
	}
}

func insertVersionSQL(driver string) string {
	switch driver {
	case "postgres":
		return `INSERT INTO schema_migrations (version) VALUES ($1)`
	default:
		return `INSERT INTO schema_migrations (version) VALUES (?)`
	}
}

// runInTx runs DDL/DML in a transaction. MySQL commits DDL implicitly; we still
// record the version after statements succeed. If recording fails, re-run may
// hit IF NOT EXISTS for DDL (idempotent migrations) or require manual fix.
func runInTx(db *sql.DB, driver string, stmts []string, fileName string, insertSQL string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	for i, st := range stmts {
		if _, err := tx.Exec(st); err != nil {
			return fmt.Errorf("statement %d: %w", i+1, err)
		}
	}

	if _, err := tx.Exec(insertSQL, fileName); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
