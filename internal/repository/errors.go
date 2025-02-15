package repository

import "errors"

var (
	ErrorInsertUser          = errors.New("insert user error")
	ErrorSelectUser          = errors.New("select user error")
	ErrorDatabase            = errors.New("database error")
	ErrorBuildingSelectQuery = errors.New("build select query error")
	ErrorBuildingInsertQuery = errors.New("build insert query error")
	ErrorUserDoesNotExist    = errors.New("user does not exist")
	ErrorUserPasswordCombine = errors.New("user password combine error")
	ErrorScanningRow         = errors.New("scanning row error")
)
