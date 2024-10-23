package usecase

import (
	"context"
	"github.com/dotrongnhan/sharing-package/pkg/logger"
	"sync"
)

type LockManager struct {
	Mu sync.Mutex
}

// NewLockManager creates a new instance of LockManager
func NewLockManager() *LockManager {
	return &LockManager{
		Mu: sync.Mutex{},
	}
}

func LockMiddleware(lm *LockManager) Middleware {
	return func(next ExecuteFunc) ExecuteFunc {
		return func(ctx context.Context, input interface{}) (interface{}, error) {
			ctxLogger := logger.NewLogger(ctx)

			// Lock the mutex
			lm.Mu.Lock()
			ctxLogger.Infof("Mutex locked")

			defer func() {
				// Unlock the mutex
				lm.Mu.Unlock()
				ctxLogger.Infof("Mutex unlocked")
			}()

			// Proceed with the next handler
			res, err := next(ctx, input)
			if err != nil {
				ctxLogger.Errorf("Error during execution: %v\n", err)
				return nil, err
			}

			return res, nil
		}
	}
}
