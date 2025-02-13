package postgres

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
)

const (
	usersTable        = "users"
	transactionsTable = "transactions"

	usernameColumn = "username"
	passwordColumn = "password"
	balanceColumn  = "balance"

	senderIDColumn   = "sender_id"
	receiverIDColumn = "receiver_id"
	amountColumn     = "amount"
	itemColumn       = "item"
	createdAtColumn  = "created_at"
)

type repository struct {
	db *DB
}

func NewRepository(db *DB) *repository {
	return &repository{db: db}
}

func (r *repository) AuthUser(ctx context.Context, user *User) error {
	builder := sq.Insert(usersTable).
		PlaceholderFormat(sq.Dollar).
		Columns(usernameColumn, passwordColumn, balanceColumn).
		Values(user.Username, user.Password, user.Coins)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("building insert query error: %w", err)
	}

	_, err = r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing insert query error: %w", err)
	}

	return nil
}

func (r *repository) TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error {
	return nil
}

func (r *repository) GetBalance(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (r *repository) GetTransactionHistory(ctx context.Context, userID string) ([]CoinTransaction, error) {
	return nil, nil
}

func (r *repository) BuyMerch(ctx context.Context, item *InventoryItem) error {
	return nil
}

func (r *repository) GetInventory(ctx context.Context, userID string) ([]InventoryItem, error) {
	return nil, nil
}

/*
func (r *repository) UpdateBalance(ctx context.Context, username string, newBalance int) error {
	builder := sq.Update(usersTable).
		PlaceholderFormat(sq.Dollar).
		Set(balanceColumn, newBalance).
		Where(sq.Eq{usernameColumn: username})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("building update query error: %w", err)
	}

	_, err = r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing update query error: %w", err)
	}

	return nil
}

func (r *repository) CreateTransaction(ctx context.Context, senderID, receiverID, amount int) error {
	builder := sq.Insert(transactionsTable).
		PlaceholderFormat(sq.Dollar).
		Columns(senderIDColumn, receiverIDColumn, amountColumn, createdAtColumn).
		Values(senderID, receiverID, amount, sq.Expr("NOW()"))

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("building insert query error: %w", err)
	}

	_, err = r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing insert query error: %w", err)
	}

	return nil
}

func (r *repository) GetUserTransactions(ctx context.Context, userID int) ([]shop.Transaction, error) {
	builder := sq.Select(senderIDColumn, receiverIDColumn, amountColumn, createdAtColumn).
		PlaceholderFormat(sq.Dollar).
		From(transactionsTable).
		Where(sq.Or{sq.Eq{senderIDColumn: userID}, sq.Eq{receiverIDColumn: userID}}).
		OrderBy(fmt.Sprintf("%s DESC", createdAtColumn))

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query error: %w", err)
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query error: %w", err)
	}
	defer rows.Close()

	var transactions []shop.Transaction
	for rows.Next() {
		var t shop.Transaction
		err := rows.Scan(&t.SenderID, &t.ReceiverID, &t.Amount, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction error: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *repository) BuyItem(ctx context.Context, userID int, item string, cost int) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction error: %w", err)
	}
	defer tx.Rollback(ctx)

	builder := sq.Select(balanceColumn).
		PlaceholderFormat(sq.Dollar).
		From(usersTable).
		Where(sq.Eq{"id": userID})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("building balance query error: %w", err)
	}

	var balance int
	err = tx.QueryRow(ctx, query, args...).Scan(&balance)
	if err != nil {
		return fmt.Errorf("fetching user balance error: %w", err)
	}
	if balance < cost {
		return errors.New("not enough coins to buy item")
	}

	builder = sq.Update(usersTable).
		PlaceholderFormat(sq.Dollar).
		Set(balanceColumn, balance-cost).
		Where(sq.Eq{"id": userID})

	query, args, err = builder.ToSql()
	if err != nil {
		return fmt.Errorf("building update balance query error: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating balance error: %w", err)
	}

	builder = sq.Insert(transactionsTable).
		PlaceholderFormat(sq.Dollar).
		Columns(senderIDColumn, itemColumn, createdAtColumn).
		Values(userID, item, time.Now())

	query, args, err = builder.ToSql()
	if err != nil {
		return fmt.Errorf("building insert transaction query error: %w", err)
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("inserting transaction error: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("committing transaction error: %w", err)
	}

	return nil
}
*/
