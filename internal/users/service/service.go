package service

import (
	"context"
	"github.com/kingxl111/merch-store/internal/users"
)

type userService struct {
	userRepo UserRepository
}

func NewUserService(usrRepo UserRepository) *userService {
	return &userService{
		userRepo: usrRepo,
	}
}

func (u *userService) Authenticate(ctx context.Context, req *users.AuthRequest) (*users.AuthResponse, error) {
	var resp users.AuthResponse
	resp.Token = "hello, token!"
	return &resp, nil
}

func (u *userService) TransferCoins(ctx context.Context, req *users.CoinTransfer) error {
	return nil
}

func (u *userService) GetUserInfo(ctx context.Context, userID string) (*users.UserInfoResponse, error) {
	return nil, nil
}
