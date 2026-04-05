package gormdb

import "gorm.io/gorm"

// WithTransaction runs fn inside a single database transaction.
//
// The *gorm.DB passed to fn is the transactional session (same concrete type as the
// root DB, but bound to one tx). Construct repositories with {DB: tx} so all work
// participates in the same transaction. If fn returns a non-nil error, GORM rolls back;
// if fn returns nil, the transaction commits.
//
// Equivalent to: return db.Transaction(fn)
func WithTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
