package service

import (
	"context"

	"github.com/kingxl111/merch-store/internal/shop"
)

type shopService struct {
	shopRepo ShopRepository
}

func NewShopService(shopRepo ShopRepository) *shopService {
	return &shopService{
		shopRepo: shopRepo,
	}
}

func (s *shopService) BuyMerch(ctx context.Context, req []shop.InventoryItem) error {
	return nil
}
