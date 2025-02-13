package service

import (
	"context"

	"github.com/kingxl111/merch-store/internal/repository/postgres"
)

type ShopRepository interface {
	BuyMerch(ctx context.Context, item *postgres.InventoryItem) error
	GetInventory(ctx context.Context, userID string) ([]postgres.InventoryItem, error)
}
