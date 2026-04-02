# Database Rules

## Migration

- SQL schema must be managed via migrations under:
  platform/db/

- Migration files must be versioned and ordered:
  0000_schema_migrations.sql
  0001_init.sql
  0002_create_post.sql
  0003_post_replies.sql
- Apply migrations with `go run ./cmd/migrate` (see `platform/db/SCHEMA.md` CLI 说明); applied filenames are stored in `schema_migrations`.

## Compatibility

- SQL must be compatible with:
  - SQLite
  - MySQL
  - PostgreSQL

- Avoid database-specific syntax

## Access Rules

- All DB access must go through repository layer
- Do not write SQL in handlers or services
- When designing database schema:
  1. First propose schema
  2. Then review against rules
  3. Then generate SQL