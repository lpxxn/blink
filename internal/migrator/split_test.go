package migrator

import (
	"strings"
	"testing"
)

func TestSplitSQLStatements_basic(t *testing.T) {
	s := `
CREATE TABLE a (id INT);
CREATE INDEX idx ON a (id);
`
	got := splitSQLStatements(s)
	if len(got) != 2 {
		t.Fatalf("want 2 statements, got %d: %q", len(got), got)
	}
}

func TestSplitSQLStatements_stringWithSemicolon(t *testing.T) {
	s := `INSERT INTO t (c) VALUES ('a;b');`
	got := splitSQLStatements(s)
	if len(got) != 1 || !strings.Contains(got[0], "a;b") {
		t.Fatalf("got %q", got)
	}
}
