package postgres

type usersRaw struct {
	id       int
	username string
	password string
	balance  int
}
