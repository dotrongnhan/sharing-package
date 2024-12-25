package sqlx_postgres

import (
	"context"
	"errors"
	sq "github.com/Masterminds/squirrel"
	"github.com/dotrongnhan/sharing-package/database"
	"github.com/dotrongnhan/sharing-package/pkg/constants"
	"github.com/dotrongnhan/sharing-package/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type repository[T any] struct {
	db    *sqlx.DB
	table string
}

func NewRepository[T any](db *sqlx.DB, table string) database.BaseRepository[T] {
	return &repository[T]{
		db:    db,
		table: table,
	}
}

func (r *repository[T]) CountByCondition(ctx context.Context, condition *database.CommonCondition) (uint64, error) {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	condition = getCondition(condition)
	db := psql.Select("count(*)").
		Where("").
		From(r.table)
	newCondition := &database.CommonCondition{
		Conditions: condition.Conditions,
		Paging:     nil,
	}
	db, err := BuildQuery(db, newCondition)
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return 0, err
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return 0, err
	}
	var total []struct {
		Count uint64 `db:"count"`
	}
	err = Select(ctx, r.db, &total, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while get total %s, err: %v", r.table, err)
		return 0, err
	}
	return total[0].Count, nil
}

func (r *repository[T]) DB() *sqlx.DB {
	return r.db
}

func (r *repository[T]) TableName() string {
	return r.table
}

func (r *repository[T]) GetByCondition(ctx context.Context, condition *database.CommonCondition) (*database.Pagination[T], error) {
	ctxLogger := logger.NewLogger(ctx)
	condition = getCondition(condition)
	total, err := r.CountByCondition(ctx, condition)
	if err != nil {
		ctxLogger.Errorf("Failed while get total, err: %v", err)
		return nil, err
	}
	meta := database.GetMetaPagination(total, condition.Paging)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	columns, err := database.GetColumnsGeneric[T]()
	db := psql.Select(columns...).
		From(r.table)
	db, err = BuildQuery(db, condition)
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	var results []*T
	err = Select(ctx, r.db, &results, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while select %s, err: %v", r.table, err)
		return nil, err
	}
	return &database.Pagination[T]{
		Data: results,
		Meta: meta,
	}, nil
}

func (r *repository[T]) GetMany(ctx context.Context, condition *database.CommonCondition) ([]*T, error) {
	ctxLogger := logger.NewLogger(ctx)
	condition = getCondition(condition)

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	columns, err := database.GetColumnsGeneric[T]()
	db := psql.Select(columns...).
		From(r.table)
	db, err = BuildQuery(db, condition)
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	var results []*T
	err = Select(ctx, r.db, &results, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while select %s, err: %v", r.table, err)
		return nil, err
	}
	return results, nil
}

func (r *repository[T]) GetById(ctx context.Context, id string) (*T, error) {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db := psql.Select("*").
		From(r.table).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil})
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	var results []*T
	err = Select(ctx, r.db, &results, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while select %s, err: %v", r.table, err)
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

func (r *repository[T]) GetByIds(ctx context.Context, ids []string) ([]*T, error) {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db := psql.Select("*").
		From(r.table).
		Where(sq.Eq{"id": ids}).
		Where(sq.Eq{"deleted_at": nil})
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	var results []*T
	err = Select(ctx, r.db, &results, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while select %s, err: %v", r.table, err)
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}

func (r *repository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	ctxLogger := logger.NewLogger(ctx)
	if createInterface, ok := any(entity).(BeforeCreateInterface); ok {
		if err := createInterface.BeforeCreate(ctx, r.db); err != nil {
			ctxLogger.Errorf("Failed while execute BeforeCreate, err: %v", err)
			return nil, err
		}
	}
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	columns, values, err := database.GetColumnsAndValues(entity)
	if err != nil {
		ctxLogger.Errorf("Failed while get columns and values, err: %v", err)
		return nil, err
	}
	db := psql.Insert(r.table).
		Columns(columns...).
		Values(values...)
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}
	id, err := Insert(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while insert %s, err: %v", r.table, err)
		return nil, err
	}
	condition := database.NewCommonCondition().WithCondition("id", id, constants.Equal)
	results, err := r.GetByCondition(ctx, condition)
	if err != nil {
		ctxLogger.Errorf("Failed while get %s by condition, err: %v", r.table, err)
		return nil, err
	}
	if results == nil || len(results.Data) == 0 {
		return nil, errors.New("not found record after insert")
	}
	return results.Data[0], nil
}

func (r *repository[T]) CreateMany(ctx context.Context, entities []*T) ([]string, error) {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	if len(entities) == 0 {
		return nil, nil
	}

	models := make([]interface{}, len(entities))
	for i, e := range entities {
		if createInterface, ok := any(e).(BeforeCreateInterface); ok {
			if err := createInterface.BeforeCreate(ctx, r.db); err != nil {
				ctxLogger.Errorf("Failed while execute BeforeCreate, err: %v", err)
				return nil, err
			}
		}
		models[i] = e
	}

	// Lấy cột từ bản ghi đầu tiên
	columns, valuesList, err := database.GetColumnsAndValuesForMany(models)
	if err != nil {
		ctxLogger.Errorf("Failed while get columns and values, err: %v", err)
		return nil, err
	}

	db := psql.Insert(r.table).
		Columns(columns...)
	for _, values := range valuesList {
		db = db.Values(values...)
	}

	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return nil, err
	}

	ids, err := InsertMultiple(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while insert %s, err: %v", r.table, err)
		return nil, err
	}

	return ids, nil
}

func (r *repository[T]) Update(ctx context.Context, id string, entity *T) error {
	ctxLogger := logger.NewLogger(ctx)
	if updateInterface, ok := any(entity).(BeforeUpdateInterface); ok {
		if err := updateInterface.BeforeUpdate(ctx, r.db); err != nil {
			ctxLogger.Errorf("Failed while execute BeforeUpdate, err: %v", err)
			return err
		}
	}
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	columns, values, err := database.GetColumnsAndValues(entity)
	if err != nil {
		ctxLogger.Errorf("Failed while get columns and values, err: %v", err)
		return err
	}
	db := psql.Update(r.table)
	for i, column := range columns {
		if column == "id" {
			continue
		}
		db = db.Set(column, values[i])
	}
	db = db.Where(sq.Eq{"id": id})
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return err
	}
	err = Exec(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while update %s, err: %v", r.table, err)
		return err
	}
	return nil
}

func (r *repository[T]) Delete(ctx context.Context, id string) error {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db := psql.Update(r.table).
		Where(sq.Eq{"id": id})
	var entity T
	if deleteInterface, ok := any(&entity).(BeforeDeleteInterface); ok {
		if err := deleteInterface.BeforeDelete(ctx, r.db, &db); err != nil {
			ctxLogger.Errorf("Failed while execute BeforeDelete, err: %v", err)
			return err
		}
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return err
	}
	err = Exec(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while delete %s, err: %v", r.table, err)
		return err
	}
	return nil
}

func (r *repository[T]) DeleteMany(ctx context.Context, ids []string) error {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db := psql.Update(r.table).
		Where(sq.Eq{"id": ids})
	var entity T
	if deleteInterface, ok := any(&entity).(BeforeDeleteInterface); ok {
		if err := deleteInterface.BeforeDelete(ctx, r.db, &db); err != nil {
			ctxLogger.Errorf("Failed while execute BeforeDelete, err: %v", err)
			return err
		}
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return err
	}
	err = Exec(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while delete %s, err: %v", r.table, err)
		return err
	}
	return nil
}

func (r *repository[T]) DeleteByCondition(ctx context.Context, condition *database.CommonCondition) error {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	condition = getCondition(condition)

	db := psql.Update(r.table)
	db, err := BuildUpdateConditions(db, condition.Conditions)
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return err
	}
	var entity T
	if deleteInterface, ok := any(&entity).(BeforeDeleteInterface); ok {
		if err = deleteInterface.BeforeDelete(ctx, r.db, &db); err != nil {
			ctxLogger.Errorf("Failed while execute BeforeDelete, err: %v", err)
			return err
		}
	}
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return err
	}
	err = Exec(ctx, r.db, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while delete %s, err: %v", r.table, err)
		return err
	}
	return nil
}

func (r *repository[T]) ExistById(ctx context.Context, id string) (bool, error) {
	ctxLogger := logger.NewLogger(ctx)
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	db := psql.Select("count(*)").
		From(r.table).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil})
	query, args, err := db.ToSql()
	if err != nil {
		ctxLogger.Errorf("Failed while build query, err: %v", err)
		return false, err
	}
	var total []struct {
		Count uint64 `db:"count"`
	}
	err = Select(ctx, r.db, &total, query, args...)
	if err != nil {
		ctxLogger.Errorf("Failed while select %s, err: %v", r.table, err)
		return false, err
	}
	return total[0].Count > 0, nil
}
