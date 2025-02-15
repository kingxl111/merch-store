package service

import (
	"context"
	"errors"

	env "github.com/kingxl111/merch-store/internal/environment"
	repo "github.com/kingxl111/merch-store/internal/repository"
	"github.com/kingxl111/merch-store/internal/repository/postgres"

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

func (s *shopService) BuyMerch(ctx context.Context, req shop.InventoryItem) error {
	username, ok := ctx.Value(env.UsernameContextKey).(string)
	if !ok {
		return shop.ErrUserNotFound
	}

	item := postgres.InventoryItem{
		Username: username,
		ItemType: req.Type,
		Quantity: req.Quantity,
	}

	err := s.shopRepo.BuyMerch(ctx, &item)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrorUserNotFound):
			return shop.ErrUserNotFound
		case errors.Is(err, repo.ErrorItemNotFound):
			return shop.ErrItemNotFound
		case errors.Is(err, repo.ErrorInsFunds):
			return shop.ErrInsufficientFunds
		case errors.Is(err, repo.ErrorBuildSenderSelectQuery),
			errors.Is(err, repo.ErrorBuildBalanceUpdateQuery),
			errors.Is(err, repo.ErrorBuildInventoryUpdateQuery):
			return shop.ErrBuildQuery
		case errors.Is(err, repo.ErrorUpdateUserBalance):
			return shop.ErrUpdateBalance
		case errors.Is(err, repo.ErrorInsertInventoryRecord):
			return shop.ErrUpdateInventory
		case errors.Is(err, repo.ErrorTxCommit),
			errors.Is(err, repo.ErrorTxBegin):
			return shop.ErrTransactionFailed
		default:
			return shop.ErrInternalError
		}
	}

	return nil
}
