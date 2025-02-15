package users

import "github.com/go-faster/errors"

var (
	ErrorService       = errors.New("users service error")
	ErrorGenerateToken = errors.New("cannot generate token")
	ErrorCreateUser    = errors.New("cannot create user")
)
