package testutil

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	"github.com/lpxxn/blink/internal/migrator"
)

// OpenSQLiteMemory runs platform migrations against an in-memory SQLite DB and returns sqlx.DB.
func OpenSQLiteMemory(t *testing.T) *sqlx.DB {
	t.Helper()
	sqldb, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })
	db := sqlx.NewDb(sqldb, "sqlite")
	dir := filepath.Join(moduleRoot(t), "platform", "db")
	if err := migrator.Run(sqldb, "sqlite", dir); err != nil {
		t.Fatal(err)
	}
	return db
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
