package postgres

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/go-faster/errors"
	"github.com/kingxl111/merch-store/internal/models"
	"time"
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

func (r *repository) CreateUser(ctx context.Context, username, password string) error {
	builder := sq.Insert(usersTable).
		PlaceholderFormat(sq.Dollar).
		Columns(usernameColumn, passwordColumn, balanceColumn).
		Values(username, password, 0)

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

func (r *repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	builder := sq.Select(usernameColumn, passwordColumn, balanceColumn).
		PlaceholderFormat(sq.Dollar).
		From(usersTable).
		Where(sq.Eq{usernameColumn: username})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query error: %w", err)
	}

	var user models.User
	err = r.db.pool.QueryRow(ctx, query, args...).Scan(&user.Username, &user.Password, &user.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found: %w", err)
	} else if err != nil {
		return nil, fmt.Errorf("executing select query error: %w", err)
	}

	return &user, nil
}

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

func (r *repository) GetUserTransactions(ctx context.Context, userID int) ([]models.Transaction, error) {
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

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
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
