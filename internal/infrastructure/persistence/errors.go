package persistence

import "errors"

const (
	UniqueViolationCode = "23505"
)

var (
	ErrNoRowsFound    = errors.New("no rows found")
	ErrNoRowsAffected = errors.New("no rows affected")
	ErrDuplicateKey   = errors.New("duplicate key")
)
