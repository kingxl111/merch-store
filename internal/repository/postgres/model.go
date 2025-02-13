package postgres

import (
	"time"
)

type User struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	Coins     int       `db:"coins"`
	CreatedAt time.Time `db:"created_at"`
}

type InventoryItem struct {
	ID       int    `db:"id"`
	UserID   string `db:"user_id"`
	ItemType string `db:"item_type"`
	Quantity int    `db:"quantity"`
}

type CoinTransaction struct {
	ID         int       `db:"id"`
	FromUserID string    `db:"from_user_id"`
	ToUserID   string    `db:"to_user_id"`
	Amount     int       `db:"amount"`
	CreatedAt  time.Time `db:"created_at"`
}

type ShopItem struct {
	ID    int    `db:"id"`
	Type  string `db:"type"`
	Price int    `db:"price"`
}

type UserInfo struct {
	User                 User
	Inventory            []InventoryItem
	SentTransactions     []CoinTransaction
	ReceivedTransactions []CoinTransaction
}
