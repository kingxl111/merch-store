package models

import "time"

type User struct {
	ID       int
	Username string
	Password string
	Balance  int
}

type Transaction struct {
	ID         int
	SenderID   int
	ReceiverID *int    // nil, если это покупка
	Item       *string // nil, если это перевод
	Amount     *int    // nil, если это покупка
	CreatedAt  time.Time
}

type MerchItem struct {
	ID    int
	Name  string
	Price int
}
