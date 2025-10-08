package internal

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/uptrace/bun"
)

var (
	ErrUnsupportedContainsType = errors.New("unsupported value type for contains comparison") // ErrInvalidContainsValue is returned when a 'contains' filter has an unsupported value type
	ErrUnsupportedBoolType     = errors.New("unsupported value type for boolean conversion")  // ErrUnsupportedBoolType is returned when a 'bool' filter receives an unsupported value type
	ErrUnsupportedNumericType  = errors.New("unsupported value type for numeric comparison")  // ErrUnsupportedNumericType is returned when a 'gt' filter receives an unsupported value type
)

// ParseFilterOperator converts a filter operator to its SQL representation. Supported operators are "AND" and "OR".
// It returns the SQL operator string and a boolean indicating if the operator is valid.
func ParseFilterOperator(op *query.FilterOperator) (string, bool) {
	return strings.ToUpper(op.Name), slices.Contains([]string{"AND", "OR"}, strings.ToUpper(op.Name))
}

// ParseFilterProperty constructs the SQL representation of a property filter. It returns a SQL condition string, SQL
// arguments array, boolean indicating if the operator is valid and an error, if any
func ParseFilterProperty(fp *query.FilterProperty) (string, []any, bool, error) {
	var condition string
	var args []any
	var err error

	switch fp.Operator {
	case "contains":
		condition, args, err = fromContains(fp.Name, fp.Value)
	case "eq":
		condition, args, err = fromEq(fp.Name, fp.Value)
	case "bool":
		condition, args, err = fromBool(fp.Name, fp.Value)
	case "gt":
		condition, args, err = fromGt(fp.Name, fp.Value)
	case "ne":
		condition, args, err = fromNe(fp.Name, fp.Value)
	default:
		return "", nil, false, nil
	}

	if err != nil {
		return "", nil, false, err
	}

	return condition, args, true, nil
}

// fromContains converts a "contains" JSON expression to an SQL expression. For strings, it uses ILIKE with '%value%'
// for case-insensitive substring matching. For arrays, it uses the @> (contains) operator to check if the column
// contains all the values in the array. Returns SQL condition string, arguments array, and error if any.
func fromContains(column string, value any) (string, []any, error) {
	switch v := value.(type) {
	case string:
		return "? ILIKE ?", []any{bun.Ident(column), "%" + v + "%"}, nil
	case []any:
		return "? @> ?", []any{bun.Ident(column), v}, nil
	}

	return "", nil, ErrUnsupportedContainsType
}

// fromEq converts an "eq" (equals) JSON expression to an SQL expression using =.
// Returns SQL condition string, arguments array, and error if any.
func fromEq(column string, value any) (string, []any, error) {
	return "? = ?", []any{bun.Ident(column), value}, nil
}

// fromBool converts a "bool" JSON expression to an SQL expression. It handles various input types (int, string, bool)
// and converts them to boolean values.
//
// - For integers: 0 is false, anything else is true
//
// - For strings: uses strconv.ParseBool
//
// - For booleans: uses the value directly
//
// Returns SQL condition string, arguments array, and error if any.
func fromBool(column string, value any) (string, []any, error) {
	var boolValue bool

	switch v := value.(type) {
	case int:
		boolValue = v != 0
	case string:
		var err error
		boolValue, err = strconv.ParseBool(v)
		if err != nil {
			return "", nil, err
		}
	case bool:
		boolValue = v
	default:
		return "", nil, ErrUnsupportedBoolType
	}

	return "? = ?", []any{bun.Ident(column), boolValue}, nil
}

// fromGt converts a "gt" (greater than) JSON expression to an SQL expression using >. It handles various numeric types
// (int, float, etc.) and string representations of numbers. For strings, it attempts to convert to int first, then to
// float if int conversion fails. Returns SQL condition string, arguments array, and error if any.
func fromGt(column string, value any) (string, []any, error) {
	switch v := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return "? > ?", []any{bun.Ident(column), v}, nil
	case string:
		var num any
		var err error

		num, err = strconv.Atoi(v)
		if err != nil {
			num, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return "", nil, err
			}
		}

		return "? > ?", []any{bun.Ident(column), num}, nil
	default:
		return "", nil, ErrUnsupportedNumericType
	}
}

// fromNe converts a "ne" (not equals) JSON expression to an SQL expression using <>. Returns SQL condition string,
// arguments array, and error if any.
func fromNe(column string, value any) (string, []any, error) {
	return "? <> ?", []any{bun.Ident(column), value}, nil
}
