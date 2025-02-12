package users

import "github.com/kingxl111/merch-store/internal/shop"

type AuthRequest struct {
	Username string
	Password string
}

type AuthResponse struct {
	Token string
}

type UserInfoResponse struct {
	Coins           int
	Inventory       []shop.InventoryItem
	ReceivedHistory []CoinTransfer
	SentHistory     []CoinTransfer
}

type CoinTransfer struct {
	User   string
	Amount int
}
