package repository

import "errors"

var (
	ErrorInsertUser = errors.New("insert user error")
	ErrorDatabase   = errors.New("database error")
)
