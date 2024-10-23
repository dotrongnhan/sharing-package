package usecase

import (
	"context"
	"sharing-package/database"
	"sharing-package/pkg/logger"
)

func TransactionMiddleware(tm database.TransactionManager) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, input interface{}) (interface{}, error) {
			ctxLogger := logger.NewLogger(ctx)
			tx := tm.GetTransaction(ctx)
			if tx != nil {
				return next(ctx, input)
			}

			ctx, err := tm.BeginTransaction(ctx)
			if err != nil {
				return nil, err
			}

			res, err := next(ctx, input)
			if err != nil {
				if rbErr := tm.RollbackTransaction(ctx); rbErr != nil {
					ctxLogger.Errorf("Failed to rollback transaction: %v\n", rbErr)
				}
				return nil, err
			}

			if err = tm.CommitTransaction(ctx); err != nil {
				ctxLogger.Errorf("Failed to commit transaction: %v\n", err)
				return nil, err
			}

			return res, nil
		}
	}
}
