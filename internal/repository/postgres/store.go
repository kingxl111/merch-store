package postgres

import (
	"context"
	"fmt"
	"time"

	repo "github.com/kingxl111/merch-store/internal/repository"

	sq "github.com/Masterminds/squirrel"
)

const (
	usersTable        = "users"
	transactionsTable = "coin_transactions"

	idColumn       = "id"
	usernameColumn = "username"
	passwordColumn = "password"
	balanceColumn  = "coins"

	senderIDColumn   = "from_user_id"
	receiverIDColumn = "to_user_id"
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

	selectBuilder := sq.Select(usernameColumn).
		PlaceholderFormat(sq.Dollar).
		From(usersTable).
		Where(sq.Eq{usernameColumn: user.Username})

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return repo.ErrorBuildingSelectQuery
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return repo.ErrorSelectUser
	}
	defer rows.Close()

	if !rows.Next() {
		builder := sq.Insert(usersTable).
			PlaceholderFormat(sq.Dollar).
			Columns(usernameColumn, passwordColumn, balanceColumn, createdAtColumn).
			Values(user.Username, user.Password, 1000, time.Now())

		query, args, err := builder.ToSql()
		if err != nil {
			return repo.ErrorBuildingInsertQuery
		}

		_, err = r.db.pool.Exec(ctx, query, args...)
		if err != nil {
			fmt.Println(query)
			fmt.Println(args...)
			return repo.ErrorInsertUser
		}
		return nil
	}

	selectBuilder = sq.Select(usernameColumn, passwordColumn).
		PlaceholderFormat(sq.Dollar).
		From(usersTable).
		Where(sq.Eq{usernameColumn: user.Username}).
		Where(sq.Eq{passwordColumn: user.Password})

	query, args, err = selectBuilder.ToSql()
	if err != nil {
		return repo.ErrorBuildingSelectQuery
	}

	rows, err = r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return repo.ErrorSelectUser
	}
	defer rows.Close()

	if !rows.Next() {
		return repo.ErrorUserPasswordCombine
	}

	return nil
}

func (r *repository) TransferCoins(ctx context.Context, fromUser, toUser string, amount int) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return repo.ErrorTxBegin
	}
	defer tx.Rollback(ctx)

	var fromUserID, fromBalance int
	var toUserID int

	selectSender := sq.Select(idColumn, balanceColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: fromUser}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(sq.Dollar)
	query, args, err := selectSender.ToSql()
	fmt.Printf("Query: %s, Args: %v\n", query, args)
	if err != nil {
		return repo.ErrorBuildSenderSelectQuery
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&fromUserID, &fromBalance)
	if err != nil {
		return repo.ErrorSenderNotFound
	}

	selectReceiver := sq.Select(idColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: toUser}).
		Suffix("FOR UPDATE").
		PlaceholderFormat(sq.Dollar)

	query, args, err = selectReceiver.ToSql()
	if err != nil {
		return repo.ErrorBuildReceiverSelectQuery
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&toUserID)
	if err != nil {
		return repo.ErrorReceiverNotFound
	}

	if fromBalance < amount {
		return repo.ErrorInsFunds
	}

	updateSender := sq.Update(usersTable).
		Set(balanceColumn, sq.Expr(balanceColumn+" - ?", amount)).
		Where(sq.Eq{idColumn: fromUserID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err = updateSender.ToSql()
	if err != nil {
		return repo.ErrorBuildSenderUpdateQuery
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return repo.ErrorUpdateSenderBalance
	}

	updateReceiver := sq.Update(usersTable).
		Set(balanceColumn, sq.Expr(balanceColumn+" + ?", amount)).
		Where(sq.Eq{idColumn: toUserID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err = updateReceiver.ToSql()
	if err != nil {
		return repo.ErrorBuildReceiverUpdateQuery
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return repo.ErrorUpdateReceiverBalance
	}

	insertTransaction := sq.Insert(transactionsTable).
		Columns(senderIDColumn, receiverIDColumn, amountColumn, createdAtColumn).
		Values(fromUserID, toUserID, amount, time.Now()).
		PlaceholderFormat(sq.Dollar)

	query, args, err = insertTransaction.ToSql()
	if err != nil {
		return repo.ErrorBuildInsertTransactionQuery
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return repo.ErrorInsertTransactionRecord
	}

	err = tx.Commit(ctx)
	if err != nil {
		return repo.ErrorTxCommit
	}

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
