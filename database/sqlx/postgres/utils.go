package sqlx_postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/dotrongnhan/sharing-package/database"
	"github.com/dotrongnhan/sharing-package/pkg/constants"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
)

func GetContextTransaction(ctx context.Context) *sql.Tx {
	if ctx.Value(constants.ContextKeyDBTransaction) != nil {
		return ctx.Value(constants.ContextKeyDBTransaction).(*sql.Tx)
	}
	return nil
}

func txSelect(tx *sql.Tx, dest interface{}, query string, args ...interface{}) error {
	rows, err := tx.Query(query, args...)
	if err != nil {
		return err
	}
	err = sqlx.StructScan(rows, dest)
	if err != nil {
		return err
	}
	return nil
}

func Select(ctx context.Context, db *sqlx.DB, dest interface{}, query string, args ...interface{}) error {
	tx := GetContextTransaction(ctx)
	if tx != nil {
		err := txSelect(tx, dest, query, args...)
		if err != nil {
			return err
		}
	} else {
		err := db.Select(dest, query, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

func Insert(ctx context.Context, db *sqlx.DB, query string, args ...interface{}) (*string, error) {
	tx := GetContextTransaction(ctx)
	queryS := fmt.Sprintf("%s %s", query, "RETURNING id")
	var id string
	var err error
	if tx != nil {
		err = tx.QueryRow(queryS, args...).Scan(&id)
	} else {
		err = db.QueryRow(queryS, args...).Scan(&id)
	}
	if err != nil {
		var mErr *mysql.MySQLError
		if errors.As(err, &mErr) {
			return nil, errors.New("data invalid")
		}
		return nil, err
	}
	return &id, nil
}

func InsertMultiple(ctx context.Context, db *sqlx.DB, query string, args ...interface{}) ([]string, error) {
	tx := GetContextTransaction(ctx)
	queryS := fmt.Sprintf("%s %s", query, "RETURNING id")
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.Query(queryS, args...)
	} else {
		rows, err = db.Query(queryS, args...)
	}
	if err != nil {
		var mErr *mysql.MySQLError
		if errors.As(err, &mErr) {
			return nil, errors.New("data invalid")
		}
		return nil, err
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func Exec(ctx context.Context, db *sqlx.DB, query string, args ...interface{}) error {
	tx := GetContextTransaction(ctx)
	var err error
	if tx != nil {
		_, err = tx.Exec(query, args...)
	} else {
		_, err = db.Exec(query, args...)
	}
	return err
}

func Delete(ctx context.Context, db *sqlx.DB, query string, args ...interface{}) error {
	tx := GetContextTransaction(ctx)
	var err error
	if tx != nil {
		_, err = tx.Exec(query, args...)
	} else {
		_, err = db.Exec(query, args...)
	}
	return err
}

func BuildQuery(db squirrel.SelectBuilder, condition *database.CommonCondition) (squirrel.SelectBuilder, error) {
	db, err := BuildConditions(db, condition.Conditions)
	if err != nil {
		return db, err
	}
	db = BuildSorting(db, condition.Sorting)
	db = BuildPaging(db, condition.Paging)
	return db, nil

}

func BuildConditions(db squirrel.SelectBuilder, conditions []database.Condition) (squirrel.SelectBuilder, error) {
	for _, cond := range conditions {
		switch strings.ToLower(cond.Op) {
		case constants.Equal:
			db = db.Where(squirrel.Eq{cond.Field: cond.Value})
		case constants.NotEqual:
			db = db.Where(squirrel.NotEq{cond.Field: cond.Value})
		case constants.LessThan:
			db = db.Where(squirrel.Lt{cond.Field: cond.Value})
		case constants.GreaterThan:
			db = db.Where(squirrel.Gt{cond.Field: cond.Value})
		case constants.LessThanOrEqual:
			db = db.Where(squirrel.LtOrEq{cond.Field: cond.Value})
		case constants.GreaterThanOrEqual:
			db = db.Where(squirrel.GtOrEq{cond.Field: cond.Value})
		case constants.In:
			db = db.Where(squirrel.Eq{cond.Field: cond.Value})
		case constants.Like:
			db = db.Where(squirrel.Like{cond.Field: cond.Value})
		case constants.NotLike:
			db = db.Where(squirrel.NotLike{cond.Field: cond.Value})
		case constants.ILike:
			db = db.Where(squirrel.ILike{cond.Field: cond.Value})
		case constants.NotILike:
			db = db.Where(squirrel.NotILike{cond.Field: cond.Value})
		default:
			return db, fmt.Errorf("unsupported operator: %s", cond.Op)
		}
	}
	return db, nil
}

func BuildSorting(db squirrel.SelectBuilder, sorting []database.Sorting) squirrel.SelectBuilder {
	for _, sort := range sorting {
		if sort.Order == constants.Asc {
			db = db.OrderBy(sort.Field)
		} else {
			db = db.OrderBy(fmt.Sprintf("%s DESC", sort.Field))
		}
	}
	return db
}

func BuildPaging(db squirrel.SelectBuilder, paging *database.Paging) squirrel.SelectBuilder {
	if paging == nil || paging.Page == 0 || paging.Limit == 0 {
		return db
	}
	limit, offset := database.GetLimitOffset(paging)
	db = db.Limit(limit).Offset(offset)
	return db
}

func BuildUpdateConditions(db squirrel.UpdateBuilder, conditions []database.Condition) (squirrel.UpdateBuilder, error) {
	for _, cond := range conditions {
		switch strings.ToLower(cond.Op) {
		case constants.Equal:
			db = db.Where(squirrel.Eq{cond.Field: cond.Value})
		case constants.NotEqual:
			db = db.Where(squirrel.NotEq{cond.Field: cond.Value})
		case constants.LessThan:
			db = db.Where(squirrel.Lt{cond.Field: cond.Value})
		case constants.GreaterThan:
			db = db.Where(squirrel.Gt{cond.Field: cond.Value})
		case constants.LessThanOrEqual:
			db = db.Where(squirrel.LtOrEq{cond.Field: cond.Value})
		case constants.GreaterThanOrEqual:
			db = db.Where(squirrel.GtOrEq{cond.Field: cond.Value})
		case constants.In:
			db = db.Where(squirrel.Eq{cond.Field: cond.Value})
		case constants.Like:
			db = db.Where(squirrel.Like{cond.Field: cond.Value})
		case constants.NotLike:
			db = db.Where(squirrel.NotLike{cond.Field: cond.Value})
		case constants.ILike:
			db = db.Where(squirrel.ILike{cond.Field: cond.Value})
		case constants.NotILike:
			db = db.Where(squirrel.NotILike{cond.Field: cond.Value})
		default:
			return db, fmt.Errorf("unsupported operator: %s", cond.Op)
		}
	}
	return db, nil
}
