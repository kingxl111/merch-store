package service

import (
	"context"
	"fmt"
	"github.com/kingxl111/merch-store/internal/shop"

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
	if req.Amount < 0 {
		return users.ErrorInvalidAmount
	}
	err := u.userRepo.TransferCoins(ctx, req.FromUser, req.ToUser, req.Amount)
	if err != nil {
		fmt.Println(err)
		if errors.Is(err, repository.ErrorInsFunds) {
			return users.ErrorInsufFunds
		}
		return users.ErrorService
	}

	return nil
}

func (u *userService) GetUserInfo(ctx context.Context, username string) (*users.UserInfoResponse, error) {
	balance, err := u.userRepo.GetBalance(ctx, username)
	if err != nil {
		return nil, users.ErrorService
	}

	inventory, err := u.userRepo.GetInventory(ctx, username)
	if err != nil {
		return nil, users.ErrorService
	}
	fmt.Println(inventory)

	transactions, err := u.userRepo.GetTransactionHistory(ctx, username)
	if err != nil {
		return nil, users.ErrorService
	}
	fmt.Println(transactions)

	var receivedHistory []users.CoinTransfer
	var sentHistory []users.CoinTransfer

	for _, tx := range transactions {
		if tx.ToUserID == username {
			receivedHistory = append(receivedHistory, users.CoinTransfer{
				FromUser: tx.FromUserID,
				ToUser:   tx.ToUserID,
				Amount:   tx.Amount,
			})
		} else if tx.FromUserID == username {
			sentHistory = append(sentHistory, users.CoinTransfer{
				FromUser: tx.FromUserID,
				ToUser:   tx.ToUserID,
				Amount:   tx.Amount,
			})
		}
	}
	userInventory := make([]shop.InventoryItem, 0)
	for _, v := range inventory {
		item := shop.InventoryItem{
			Type:     v.ItemType,
			Quantity: v.Quantity,
		}
		userInventory = append(userInventory, item)
	}

	resp := &users.UserInfoResponse{
		Coins:           *balance,
		Inventory:       userInventory,
		ReceivedHistory: receivedHistory,
		SentHistory:     sentHistory,
	}

	return resp, nil
}
