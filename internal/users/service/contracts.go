package service

import (
	"context"

	"github.com/kingxl111/merch-store/internal/repository/postgres"
)

type AuthRepository interface {
	AuthUser(ctx context.Context, user *postgres.User) error
}

type UserRepository interface {
	GetInventory(ctx context.Context, username string) ([]postgres.InventoryItem, error)
	TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error
	GetBalance(ctx context.Context, username string) (*int, error)
	GetTransactionHistory(ctx context.Context, username string) ([]postgres.CoinTransaction, error)
}
