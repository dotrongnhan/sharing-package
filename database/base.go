package database

import (
	"context"
)

type BaseRepository[T any] interface {
	CountByCondition(ctx context.Context, condition *CommonCondition) (uint64, error)
	GetByCondition(ctx context.Context, condition *CommonCondition) (*Pagination[T], error)
	GetMany(ctx context.Context, condition *CommonCondition) ([]*T, error)
	GetById(ctx context.Context, id string) (*T, error)
	GetByIds(ctx context.Context, ids []string) ([]*T, error)
	Create(ctx context.Context, entity *T) (*T, error)
	CreateMany(ctx context.Context, entity []*T) ([]string, error)
	Update(ctx context.Context, id string, entity *T) error
	Delete(ctx context.Context, id string) error
	DeleteByCondition(ctx context.Context, condition *CommonCondition) error
	DeleteMany(ctx context.Context, ids []string) error
	ExistById(ctx context.Context, id string) (bool, error)
}
