package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dotrongnhan/sharing-package/pkg/constants"
	"github.com/jmoiron/sqlx"
)

type transactionManager struct {
	db *sqlx.DB
}

type TransactionManager interface {
	BeginTransaction(ctx context.Context) (context.Context, error)
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context) error
	GetTransaction(ctx context.Context) interface{}
}

func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &transactionManager{db: db}
}

func getContextTransaction(ctx context.Context) *sql.Tx {
	if ctx.Value(constants.ContextKeyDBTransaction) != nil {
		return ctx.Value(constants.ContextKeyDBTransaction).(*sql.Tx)
	}
	return nil
}

func (tm *transactionManager) BeginTransaction(ctx context.Context) (context.Context, error) {
	tx := getContextTransaction(ctx)
	if tx != nil {
		return ctx, nil
	}

	tx, err := tm.db.Begin()
	if err != nil {
		return ctx, err
	}

	ctx = context.WithValue(ctx, constants.ContextKeyDBTransaction, tx)
	return ctx, nil
}

func (tm *transactionManager) CommitTransaction(ctx context.Context) error {
	tx := getContextTransaction(ctx)
	if tx == nil {
		return errors.New("no transaction found in context")
	}
	return tx.Commit()
}

func (tm *transactionManager) RollbackTransaction(ctx context.Context) error {
	tx := getContextTransaction(ctx)
	if tx == nil {
		return errors.New("no transaction found in context")
	}
	return tx.Rollback()
}

func (tm *transactionManager) GetTransaction(ctx context.Context) interface{} {
	return ctx.Value(constants.ContextKeyDBTransaction)
}
