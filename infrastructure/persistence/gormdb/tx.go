package gormdb

import "gorm.io/gorm"

// RunInTransaction runs fn inside a single GORM transaction.
// Inside fn, use the provided tx (not the root db) for every repository so all operations
// share the same session and commit or roll back together.
func RunInTransaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// Repos holds repositories bound to one *gorm.DB — either the root connection or a transactional session.
type Repos struct {
	Users *UserRepository
	OAuth *OAuthRepository
}

// NewRepos returns repositories that all use the same db handle (root or tx).
func NewRepos(db *gorm.DB) *Repos {
	return &Repos{
		Users: &UserRepository{DB: db},
		OAuth: &OAuthRepository{DB: db},
	}
}
