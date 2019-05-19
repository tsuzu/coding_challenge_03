package model

import "errors"

var (
	// ErrNoUser means there is no target user in db
	ErrNoUser = errors.New("specified user is not found")
)
