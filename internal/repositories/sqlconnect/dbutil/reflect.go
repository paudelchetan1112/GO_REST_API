package dbutil

import (
	
	"reflect"
	"strings"
)

// GetStructValues extracts exported struct field values whose db tags
// are non-empty and are not ignored using db:"-".
// It returns nil for nil or non-struct inputs.
func GetStructValues(model interface{}) []interface{} {
	if model == nil {
		return nil
	}

	v := reflect.ValueOf(model)

	// Dereference pointers safely.
	for v.IsValid() && v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	values := make([]interface{}, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			continue
		}

		parts := strings.Split(dbTag, ",")
		columnName := strings.TrimSpace(parts[0])

		if columnName == "" || columnName == "-" {
			continue
		}

		values = append(values, fieldValue.Interface())
	}

	return values
}




