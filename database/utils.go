package database

import (
	"errors"
	"reflect"
)

func GetLimitOffset(paging *Paging) (uint64, uint64) {
	if paging == nil {
		return 0, 0
	}
	if paging.Limit == 0 {
		paging.Limit = 10
	}
	if paging.Page == 0 {
		paging.Page = 1
	}
	offset := (paging.Page - 1) * paging.Limit
	return paging.Limit, offset
}

func GetColumns(model interface{}) ([]string, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, errors.New("model must be a struct")
	}

	var columns []string
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Struct && field.Type.Name() == "Base" {
			// Gọi đệ quy để lấy các trường từ `Base`
			baseColumns, err := GetColumns(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			// Thêm các trường của Base vào array
			columns = append(columns, baseColumns...)
			continue
		}
		columnName := field.Tag.Get("db") // Giả sử bạn có tag `db` trong struct để chỉ định tên cột
		if columnName == "" {
			columnName = field.Name // Sử dụng tên trường nếu không có tag `db`
		}
		columns = append(columns, columnName)
	}
	return columns, nil
}

func GetColumnsGeneric[T any]() ([]string, error) {
	// Tạo một instance của T để truyền vào GetColumns
	var model T
	return GetColumns(model)
}

func GetColumnsAndValues(model interface{}) ([]string, []interface{}, error) {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, nil, errors.New("model must be a struct")
	}

	var columns []string
	var values []interface{}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		if field.Tag.Get("omit") != "" && fieldValue.IsZero() {
			continue
		}
		if fieldValue.Kind() == reflect.Struct && field.Type.Name() == "Base" {
			baseColumns, baseValues, err := GetColumnsAndValues(fieldValue.Interface())
			if err != nil {
				return nil, nil, err
			}
			columns = append(columns, baseColumns...)
			values = append(values, baseValues...)
			continue
		}
		columnName := field.Tag.Get("db") // Giả sử bạn có tag `db` trong struct để chỉ định tên cột
		if columnName == "" {
			columnName = field.Name // Sử dụng tên trường nếu không có tag `db`
		}

		columns = append(columns, columnName)
		values = append(values, v.Field(i).Interface())
	}
	return columns, values, nil
}

func GetColumnsAndValuesForMany(models []interface{}) ([]string, [][]interface{}, error) {
	values := make([][]interface{}, 0, len(models))
	columns, value, err := GetColumnsAndValues(models[0])
	if err != nil {
		return nil, nil, err
	}
	values = append(values, value)
	if len(models) == 1 {
		return columns, values, nil
	}
	for i := 1; i < len(models); i++ {
		_, newValues, err2 := GetColumnsAndValues(models[i])
		if err2 != nil {
			return nil, nil, err2
		}
		values = append(values, newValues)
	}
	return columns, values, nil
}

func GetMetaPagination(total uint64, paging *Paging) *Meta {
	if total == 0 || paging == nil || paging.Limit == 0 {
		return &Meta{
			ItemsPerPage: total,
			TotalItems:   total,
			CurrentPage:  1,
			TotalPages:   1,
		}
	}

	totalPages := (total + paging.Limit - 1) / paging.Limit
	return &Meta{
		ItemsPerPage: paging.Limit,
		TotalItems:   total,
		CurrentPage:  paging.Page,
		TotalPages:   totalPages,
	}
}
