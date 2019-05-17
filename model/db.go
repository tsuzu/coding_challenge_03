package model

import "database/sql"

// DB interface represents sql.DB or sql.Tx etc...
type DB interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}
