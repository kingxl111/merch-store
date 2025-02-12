package shop

type SendCoinRequest struct {
	ToUser string
	Amount int
}

type InventoryItem struct {
	Type     string
	Quantity int
}
