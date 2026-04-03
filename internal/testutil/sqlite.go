package testutil

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/lpxxn/blink/internal/migrator"
)

// OpenSQLiteMemory runs platform migrations against an in-memory SQLite DB.
func OpenSQLiteMemory(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	dir := filepath.Join(moduleRoot(t), "platform", "db")
	if err := migrator.Run(db, "sqlite", dir); err != nil {
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
