package mongo

import (
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// toBSONOmitZero converts a struct to bson.M, omitting zero-value fields.
// This mimics PostgreSQL's OmitZero() behavior for consistency across backends.
func toBSONOmitZero(v interface{}) bson.M {
	result := bson.M{}
	val := reflect.ValueOf(v)

	// Handle pointer
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return result
		}
		val = val.Elem()
	}

	// Only works with structs
	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get bson tag
		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		// Parse bson tag (e.g., "field_name,omitempty")
		fieldName := bsonTag
		for idx := 0; idx < len(bsonTag); idx++ {
			if bsonTag[idx] == ',' {
				fieldName = bsonTag[:idx]

				break
			}
		}

		// Skip if zero value
		if isZeroValue(field) {
			continue
		}

		result[fieldName] = field.Interface()
	}

	return result
}

// isZeroValue checks if a reflect.Value is a zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// Special handling for time.Time
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// For other structs, check if all fields are zero
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}

		return true
	}

	return false
}
