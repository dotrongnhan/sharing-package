package usecase

import "context"

type Middleware func(ExecuteFunc) ExecuteFunc

type ExecuteFunc func(ctx context.Context, input interface{}) (interface{}, error)

// WrapMiddlewares wraps the given middlewares around an ExecuteFunc
func WrapMiddlewares(execute ExecuteFunc, middlewares ...Middleware) ExecuteFunc {
	for _, m := range middlewares {
		execute = m(execute)
	}
	return execute
}

type WithMiddleware struct {
	ExecuteFunc ExecuteFunc
}

func (u *WithMiddleware) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	return u.ExecuteFunc(ctx, input)
}

func NewWithMiddleware(execute ExecuteFunc, middlewares ...Middleware) *WithMiddleware {
	return &WithMiddleware{
		ExecuteFunc: WrapMiddlewares(execute, middlewares...),
	}
}
