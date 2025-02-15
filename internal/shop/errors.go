package shop

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrItemNotFound      = errors.New("item not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrBuildQuery        = errors.New("failed to build query")
	ErrUpdateBalance     = errors.New("failed to update user balance")
	ErrUpdateInventory   = errors.New("failed to update inventory")
	ErrTransactionFailed = errors.New("transaction failed")
	ErrInternalError     = errors.New("internal error")
)
