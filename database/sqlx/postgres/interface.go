package sqlx_postgres

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type DatabaseAccessor interface {
	DB() *sqlx.DB
	TableName() string
}

type BeforeCreateInterface interface {
	BeforeCreate(context.Context, *sqlx.DB) error
}

type BeforeUpdateInterface interface {
	BeforeUpdate(context.Context, *sqlx.DB) error
}

type BeforeDeleteInterface interface {
	BeforeDelete(context.Context, *sqlx.DB, *squirrel.UpdateBuilder) error
}
