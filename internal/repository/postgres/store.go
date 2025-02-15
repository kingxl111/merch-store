package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	repo "github.com/kingxl111/merch-store/internal/repository"

	sq "github.com/Masterminds/squirrel"
)

const (
	usersTable        = "users"
	transactionsTable = "coin_transactions"
	shopItemsTable    = "shop_items"
	inventoryTable    = "inventory"

	idColumn       = "id"
	usernameColumn = "username"
	passwordColumn = "password"
	balanceColumn  = "coins"

	senderIDColumn   = "from_user_id"
	receiverIDColumn = "to_user_id"
	amountColumn     = "amount"
	itemColumn       = "item"
	createdAtColumn  = "created_at"

	priceColumn = "price"
	typeColumn  = "type"

	userIDColumn   = "user_id"
	itemTypeColumn = "item_type"
	quantityColumn = "quantity"
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
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.Error("cannot rollback transaction", slog.Any("error", rollbackErr))
			}
		}
	}()

	var fromUserID, toUserID uuid.UUID
	var fromBalance int

	selectSender := sq.Select(idColumn, balanceColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: fromUser}).
		Suffix("FOR UPDATE SKIP LOCKED").
		PlaceholderFormat(sq.Dollar)
	query, args, err := selectSender.ToSql()
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
		Suffix("FOR UPDATE SKIP LOCKED").
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

func (r *repository) GetBalance(ctx context.Context, username string) (*int, error) {
	builder := sq.Select(balanceColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: username}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, repo.ErrorBuildBalanceSelectQuery
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, repo.ErrorSelectBalance
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, repo.ErrorUserNotFound
	}

	var balance int
	err = rows.Scan(&balance)
	if err != nil {
		return nil, repo.ErrorScanBalance
	}

	return &balance, nil
}

func (r *repository) GetTransactionHistory(ctx context.Context, userID string) ([]CoinTransaction, error) {
	builder := sq.Select(idColumn, senderIDColumn, receiverIDColumn, amountColumn, createdAtColumn).
		From(transactionsTable).
		Where(sq.Or{
			sq.Eq{senderIDColumn: userID},
			sq.Eq{receiverIDColumn: userID},
		}).
		OrderBy(createdAtColumn + " DESC").
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, repo.ErrorBuildTransactionQuery
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, repo.ErrorSelectTransactions
	}
	defer rows.Close()

	var transactions []CoinTransaction
	for rows.Next() {
		var t CoinTransaction
		if err := rows.Scan(&t.ID, &t.FromUserID, &t.ToUserID, &t.Amount, &t.CreatedAt); err != nil {
			return nil, repo.ErrorScanTransaction
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}
func (r *repository) BuyMerch(ctx context.Context, item *InventoryItem) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return repo.ErrorTxBegin
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.Error("cannot rollback transaction", slog.Any("error", rollbackErr))
			}
		}
	}()

	var userID uuid.UUID
	var balance int

	selectBalance := sq.Select(idColumn, balanceColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: item.Username}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectBalance.ToSql()
	if err != nil {
		return repo.ErrorBuildSenderSelectQuery
	}

	err = r.db.pool.QueryRow(ctx, query, args...).Scan(&userID, &balance)
	if err != nil {
		return repo.ErrorUserNotFound
	}

	var price int
	selectItem := sq.Select(priceColumn).
		From(shopItemsTable).
		Where(sq.Eq{typeColumn: item.ItemType}).
		PlaceholderFormat(sq.Dollar)

	query, args, err = selectItem.ToSql()
	if err != nil {
		return repo.ErrorBuildSenderSelectQuery
	}

	err = r.db.pool.QueryRow(ctx, query, args...).Scan(&price)
	if err != nil {
		return repo.ErrorItemNotFound
	}

	totalCost := price * item.Quantity

	if balance < totalCost {
		return repo.ErrorInsFunds
	}

	selectUserForUpdate := sq.Select(idColumn).
		From(usersTable).
		Where(sq.Eq{usernameColumn: item.Username}).
		Suffix("FOR UPDATE SKIP LOCKED").
		PlaceholderFormat(sq.Dollar)

	query, args, err = selectUserForUpdate.ToSql()
	if err != nil {
		return repo.ErrorBuildSenderSelectQuery
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		return repo.ErrorUserNotFound
	}

	updateBalance := sq.Update(usersTable).
		Set(balanceColumn, sq.Expr(balanceColumn+" - ?", totalCost)).
		Where(sq.Eq{idColumn: userID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err = updateBalance.ToSql()
	if err != nil {
		return repo.ErrorBuildBalanceUpdateQuery
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return repo.ErrorUpdateUserBalance
	}

	upsertInventory := sq.Insert(inventoryTable).
		Columns(userIDColumn, itemTypeColumn, quantityColumn).
		Values(userID, item.ItemType, item.Quantity).
		Suffix("ON CONFLICT (user_id, item_type) DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity").
		PlaceholderFormat(sq.Dollar)

	query, args, err = upsertInventory.ToSql()
	if err != nil {
		return repo.ErrorBuildInventoryUpdateQuery
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return repo.ErrorInsertInventoryRecord
	}

	err = tx.Commit(ctx)
	if err != nil {
		return repo.ErrorTxCommit
	}
	return nil
}

func (r *repository) GetInventory(ctx context.Context, username string) ([]InventoryItem, error) {
	builder := sq.Select("i.id", "i.user_id", "i.item_type", "i.quantity").
		From("inventory i").
		Join("users u ON u.id = i.user_id").
		Where(sq.Eq{"u.username": username}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, repo.ErrorBuildInventorySelectQuery
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, repo.ErrorSelectInventory
	}
	defer rows.Close()

	var inventory []InventoryItem
	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(&item.ID, &item.UserID, &item.ItemType, &item.Quantity); err != nil {
			return nil, repo.ErrorScanQuery
		}
		inventory = append(inventory, item)
	}
	return inventory, nil
}
