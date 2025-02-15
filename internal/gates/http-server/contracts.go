package http_server

import (
	"context"

	"github.com/kingxl111/merch-store/internal/shop"
	"github.com/kingxl111/merch-store/internal/users"
)

//go:generate mockgen -source=contracts.go -destination=mocks.go -package=http_server
type (
	UserService interface {
		Authenticate(ctx context.Context, req *users.AuthRequest) (*users.AuthResponse, error)
		TransferCoins(ctx context.Context, req *users.CoinTransfer) error
		GetUserInfo(ctx context.Context, username string) (*users.UserInfoResponse, error)
	}

	ShopService interface {
		BuyMerch(ctx context.Context, req shop.InventoryItem) error
	}
)
