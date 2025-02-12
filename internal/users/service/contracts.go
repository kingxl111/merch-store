package service

import (
	"context"
	pg "github.com/kingxl111/merch-store/internal/repository/postgres"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *pg.User) error
	GetUserByCredentials(ctx context.Context, username, password string) (*pg.User, error)
}

type CoinRepository interface {
	GetBalance(ctx context.Context, userID string) (int, error)
	TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error
	GetTransactionHistory(ctx context.Context, userID string) (*pg.CoinHistory, error)
}

type InventoryRepository interface {
	GetInventory(ctx context.Context, userID string) ([]pg.InventoryItem, error)
	AddItem(ctx context.Context, userID string, item *pg.InventoryItem) error
}

type ShopRepository interface {
	GetItemPrice(ctx context.Context, itemType string) (int, error)
	UpdateStock(ctx context.Context, itemType string, quantity int) error
}
