package service

import (
	"context"
	"github.com/go-faster/errors"
	"github.com/kingxl111/merch-store/internal/repository"
	"github.com/kingxl111/merch-store/internal/repository/postgres"
	"github.com/kingxl111/merch-store/internal/users"
)

type userService struct {
	userRepo UserRepository
	authRepo AuthRepository
}

func NewUserService(usrRepo UserRepository, authRepo AuthRepository) *userService {
	return &userService{
		userRepo: usrRepo,
		authRepo: authRepo,
	}
}

func (u *userService) Authenticate(ctx context.Context, req *users.AuthRequest) (*users.AuthResponse, error) {
	var resp users.AuthResponse
	usrRepo := postgres.User{
		Username: req.Username,
		Password: req.Password,
	}
	usrRepo.Password = generatePasswordHash(usrRepo.Password)

	err := u.authRepo.AuthUser(ctx, &usrRepo)
	if err != nil {
		if errors.Is(err, repository.ErrorInsertUser) {
			return nil, users.ErrorCreateUser
		}
		if errors.Is(err, repository.ErrorUserPasswordCombine) {
			return nil, users.ErrorWrongPassword
		}
		return nil, users.ErrorService
	}

	token, err := GenerateToken(req.Username)
	if err != nil {
		return nil, users.ErrorGenerateToken
	}
	resp.Token = token
	return &resp, nil
}

func (u *userService) TransferCoins(ctx context.Context, req *users.CoinTransfer) error {
	return nil
}

func (u *userService) GetUserInfo(ctx context.Context, userID string) (*users.UserInfoResponse, error) {
	return nil, nil
}
