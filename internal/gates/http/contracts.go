package http

import (
	"context"
	model "github.com/kingxl111/merch-store/internal/users"
)

type AuthService interface {
	Authenticate(ctx context.Context, req *model.AuthRequest) (*model.AuthResponse, error)
	CreateUser(ctx context.Context, req *model.AuthRequest) (*model.AuthResponse, error)
}

type UserInfoService interface {
	GetUserInfo(ctx context.Context, userID string) (*model.UserInfoResponse, error)
	GetTransactionHistory(ctx context.Context, userID string) (*model.CoinHistoryResponse, error)
}

type TokenService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(token string) (string, error)
}
