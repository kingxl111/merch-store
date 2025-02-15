package repository

import "errors"

var (
	ErrorInsertUser          = errors.New("insert user error")
	ErrorSelectUser          = errors.New("select user error")
	ErrorDatabase            = errors.New("database error")
	ErrorBuildingSelectQuery = errors.New("build select query error")
	ErrorBuildingInsertQuery = errors.New("build insert query error")
	ErrorUserPasswordCombine = errors.New("user password combine error")

	ErrorInsFunds = errors.New("insufficient funds")
	ErrorTxBegin  = errors.New("failed to begin transaction")
	ErrorTxCommit = errors.New("failed to commit transaction")

	ErrorBuildSenderSelectQuery   = errors.New("failed to build sender select query")
	ErrorSenderNotFound           = errors.New("sender not found")
	ErrorBuildReceiverSelectQuery = errors.New("failed to build receiver select query")
	ErrorReceiverNotFound         = errors.New("receiver not found")

	ErrorBuildSenderUpdateQuery = errors.New("failed to build sender update query")
	ErrorUpdateSenderBalance    = errors.New("failed to update sender balance")

	ErrorBuildReceiverUpdateQuery = errors.New("failed to build receiver update query")
	ErrorUpdateReceiverBalance    = errors.New("failed to update receiver balance")

	ErrorBuildInsertTransactionQuery = errors.New("failed to build insert transaction query")
	ErrorInsertTransactionRecord     = errors.New("failed to insert transaction record")
)
