package testutil

import (
	"path/filepath"
	"runtime"
	"testing"

	glsqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/lpxxn/blink/internal/migrator"
)

// OpenSQLiteMemory runs platform migrations against an in-memory SQLite DB and returns *gorm.DB.
func OpenSQLiteMemory(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := gorm.Open(glsqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqldb, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })
	sqldb.SetMaxOpenConns(1)
	dir := filepath.Join(moduleRoot(t), "platform", "db")
	if err := migrator.Run(sqldb, "sqlite", dir); err != nil {
		t.Fatal(err)
	}
	return gdb
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
