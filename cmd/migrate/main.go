// Command migrate creates/opens a database and applies SQL files from -dir in order,
// recording each applied filename in schema_migrations (see 0000_schema_migrations.sql).
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lpxxn/blink/internal/migrator"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// usageText is shown with -h / -help (see flag.Usage). Full guide: cmd/migrate/README.md
const usageText = `blink migrate — 按顺序执行尚未应用的 *.sql，并将文件名记入 schema_migrations。

详细说明（SQLite / PostgreSQL 示例、参数、FAQ）见仓库内文档:
  cmd/migrate/README.md

用法（在仓库根目录）:
  go run ./cmd/migrate [-driver DRIVER] [-dsn DSN] [-dir DIR]

速览:
  • 默认: sqlite + file:blink.db?cache=shared&mode=rwc + -dir platform/db
  • 请在仓库根执行，或显式 -dir 指向含 0000_schema_migrations.sql 的目录

参数:
`

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
		flag.PrintDefaults()
	}

	driverFlag := flag.String("driver", "sqlite", `database driver: "sqlite", "mysql", or "postgres"`)
	dsnFlag := flag.String("dsn", "file:blink.db?cache=shared&mode=rwc", "data source name (driver-specific)")
	dirFlag := flag.String("dir", "platform/db", "directory containing ordered *.sql migrations")
	flag.Parse()

	if err := run(*driverFlag, *dsnFlag, *dirFlag); err != nil {
		log.Fatal(err)
	}
}

func run(driver, dsn, dir string) error {
	d, err := normalizeDriver(driver)
	if err != nil {
		return err
	}

	db, err := sql.Open(d, dsn)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	// Single connection avoids SQLite locking surprises during migrations.
	db.SetMaxOpenConns(1)

	fmt.Fprintf(os.Stderr, "driver=%s dsn=%s dir=%s\n", d, redactDSN(d, dsn), dir)

	return migrator.Run(db, d, dir)
}

func normalizeDriver(driver string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "sqlite", "sqlite3":
		return "sqlite", nil
	case "mysql":
		return "mysql", nil
	case "postgres", "postgresql", "pg":
		return "postgres", nil
	default:
		return "", fmt.Errorf("unknown -driver %q (use sqlite, mysql, postgres)", driver)
	}
}

func redactDSN(d, dsn string) string {
	if d != "mysql" && d != "postgres" {
		return dsn
	}
	if i := strings.IndexByte(dsn, '@'); i > 0 {
		head := dsn[:i]
		if j := strings.LastIndexByte(head, ':'); j >= 0 {
			return head[:j+1] + "***" + dsn[i:]
		}
	}
	return dsn
}
