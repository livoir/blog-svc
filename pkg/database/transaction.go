package database

import (
	"database/sql"
	"errors"
	"livoir-blog/internal/domain"
)

type SQLTransaction struct {
	tx *sql.Tx
}

func (t *SQLTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *SQLTransaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *SQLTransaction) GetTx() *sql.Tx {
	return t.tx
}

type SQLTransactor struct {
	db *sql.DB
}

func NewSQLTransactor(db *sql.DB) (domain.Transactor, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}
	return &SQLTransactor{db: db}, nil
}

func (s *SQLTransactor) BeginTx() (domain.Transaction, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	return &SQLTransaction{tx: tx}, nil
}
