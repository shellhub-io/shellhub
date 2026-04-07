package internal

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/uptrace/bun"
)

var (
	ErrUnsupportedContainsType = errors.New("unsupported value type for contains comparison") // ErrInvalidContainsValue is returned when a 'contains' filter has an unsupported value type
	ErrUnsupportedBoolType     = errors.New("unsupported value type for boolean conversion")  // ErrUnsupportedBoolType is returned when a 'bool' filter receives an unsupported value type
	ErrUnsupportedNumericType  = errors.New("unsupported value type for numeric comparison")  // ErrUnsupportedNumericType is returned when a 'gt' filter receives an unsupported value type
)

// qualifyColumn returns a bun.Ident for the given column, optionally prefixed
// with a table alias (e.g. "device.name") to avoid ambiguity in JOINed queries.
func qualifyColumn(column, tableAlias string) bun.Ident {
	if tableAlias != "" {
		return bun.Ident(tableAlias + "." + column)
	}

	return bun.Ident(column)
}

// TODO: remove when MongoDB support is dropped.
// Maps Mongo-style paths (e.g. "info.platform") to Postgres columns ("platform").
var legacyMongoFieldMapping = map[string]string{
	"info.platform":      "platform",
	"info.id":            "identifier",
	"info.pretty_name":   "pretty_name",
	"info.version":       "version",
	"info.arch":          "arch",
	"identity.mac":       "mac",
	"position.longitude": "longitude",
	"position.latitude":  "latitude",
}

// mapColumnFromLegacyMongo translates MongoDB-style nested field paths to PostgreSQL column names.
// See legacyMongoFieldMapping for details. Returns the original column if no mapping exists.
func mapColumnFromLegacyMongo(column string) string {
	if mapped, ok := legacyMongoFieldMapping[column]; ok {
		return mapped
	}

	return column
}

// TODO: remove when MongoDB support is dropped.
// "online" is a virtual field (not a real column), so we can't filter by it directly in WHERE.
// We expand it to the actual expression: disconnected_at IS NULL AND last_seen > (now - 2min).
func fromOnlineFilter(value any) (string, []any, bool, error) {
	var isOnline bool

	switch v := value.(type) {
	case bool:
		isOnline = v
	case string:
		var err error
		isOnline, err = strconv.ParseBool(v)
		if err != nil {
			return "", nil, false, err
		}
	default:
		return "", nil, false, ErrUnsupportedBoolType
	}

	threshold := time.Now().Add(-2 * time.Minute)

	if isOnline {
		return `("device"."disconnected_at" IS NULL AND "device"."last_seen" > ?)`, []any{threshold}, true, nil
	}

	return `("device"."disconnected_at" IS NOT NULL OR "device"."last_seen" <= ?)`, []any{threshold}, true, nil
}

// ParseFilterOperator converts a filter operator to its SQL representation. Supported operators are "AND" and "OR".
// It returns the SQL operator string and a boolean indicating if the operator is valid.
func ParseFilterOperator(op *query.FilterOperator) (string, bool) {
	return strings.ToUpper(op.Name), slices.Contains([]string{"AND", "OR"}, strings.ToUpper(op.Name))
}

// ParseFilterProperty constructs the SQL representation of a property filter.
// tableAlias, when non-empty, qualifies column names to avoid ambiguity in
// queries with JOINs (e.g. "device.name" instead of just "name").
// It returns a SQL condition string, SQL arguments array, boolean indicating
// if the operator is valid and an error, if any.
func ParseFilterProperty(fp *query.FilterProperty, tableAlias string) (string, []any, bool, error) {
	// Handle virtual fields that don't exist as real columns (see fromOnlineFilter for details)
	if fp.Name == "online" {
		return fromOnlineFilter(fp.Value)
	}

	// tags.name requires an EXISTS subquery through the device_tags junction table,
	// because tags live in a separate table with a many-to-many relationship.
	if fp.Name == "tags.name" {
		return fromTagsFilter(fp.Operator, fp.Value)
	}

	var condition string
	var args []any
	var err error

	switch fp.Operator {
	case "contains":
		condition, args, err = fromContains(fp.Name, fp.Value, tableAlias)
	case "eq":
		condition, args, err = fromEq(fp.Name, fp.Value, tableAlias)
	case "bool":
		condition, args, err = fromBool(fp.Name, fp.Value, tableAlias)
	case "gt":
		condition, args, err = fromGt(fp.Name, fp.Value, tableAlias)
	case "ne":
		condition, args, err = fromNe(fp.Name, fp.Value, tableAlias)
	default:
		return "", nil, false, nil
	}

	if err != nil {
		return "", nil, false, err
	}

	return condition, args, true, nil
}

// fromTagsFilter handles "tags.name" filters by generating an EXISTS subquery
// through the device_tags junction table. For "contains" with a string value, it
// matches tag names using ILIKE. For "contains" with an array value, it checks
// that the device has all specified tags. For "eq", it checks for an exact tag name.
func fromTagsFilter(operator string, value any) (string, []any, bool, error) {
	const base = `EXISTS (SELECT 1 FROM "device_tags" JOIN "tags" ON "tags"."id" = "device_tags"."tag_id" WHERE "device_tags"."device_id" = "device"."id" AND `

	switch operator {
	case "contains":
		switch v := value.(type) {
		case string:
			return base + `"tags"."name" ILIKE ?)`, []any{"%" + v + "%"}, true, nil
		case []any:
			strs := make([]string, len(v))
			for i, item := range v {
				s, ok := item.(string)
				if !ok {
					return "", nil, false, ErrUnsupportedContainsType
				}
				strs[i] = s
			}

			// Use a counting subquery to ensure AND semantics: the device must have ALL
			// specified tags, consistent with MongoDB's $all and the generic PG @> operator.
			return `(SELECT COUNT(DISTINCT "tags"."name") FROM "device_tags" JOIN "tags" ON "tags"."id" = "device_tags"."tag_id" WHERE "device_tags"."device_id" = "device"."id" AND "tags"."name" IN (?)) = ?`,
				[]any{bun.List(strs), len(strs)}, true, nil
		default:
			return "", nil, false, ErrUnsupportedContainsType
		}
	case "eq":
		return base + `"tags"."name" = ?)`, []any{value}, true, nil
	default:
		return "", nil, false, nil
	}
}

// fromContains converts a "contains" JSON expression to an SQL expression. For strings, it uses ILIKE with '%value%'
// for case-insensitive substring matching. For arrays, it uses the @> (contains) operator to check if the column
// contains all the values in the array. Returns SQL condition string, arguments array, and error if any.
func fromContains(column string, value any, tableAlias string) (string, []any, error) {
	column = mapColumnFromLegacyMongo(column)

	switch v := value.(type) {
	case string:
		return "? ILIKE ?", []any{qualifyColumn(column, tableAlias), "%" + v + "%"}, nil
	case []any:
		return "? @> ?", []any{qualifyColumn(column, tableAlias), v}, nil
	}

	return "", nil, ErrUnsupportedContainsType
}

// fromEq converts an "eq" (equals) JSON expression to an SQL expression using =.
// Returns SQL condition string, arguments array, and error if any.
func fromEq(column string, value any, tableAlias string) (string, []any, error) {
	return "? = ?", []any{qualifyColumn(mapColumnFromLegacyMongo(column), tableAlias), value}, nil
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
func fromBool(column string, value any, tableAlias string) (string, []any, error) {
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

	return "? = ?", []any{qualifyColumn(mapColumnFromLegacyMongo(column), tableAlias), boolValue}, nil
}

// fromGt converts a "gt" (greater than) JSON expression to an SQL expression using >. It handles various numeric types
// (int, float, etc.) and string representations of numbers. For strings, it attempts to convert to int first, then to
// float if int conversion fails. Returns SQL condition string, arguments array, and error if any.
func fromGt(column string, value any, tableAlias string) (string, []any, error) {
	column = mapColumnFromLegacyMongo(column)

	switch v := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return "? > ?", []any{qualifyColumn(column, tableAlias), v}, nil
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

		return "? > ?", []any{qualifyColumn(column, tableAlias), num}, nil
	default:
		return "", nil, ErrUnsupportedNumericType
	}
}

// fromNe converts a "ne" (not equals) JSON expression to an SQL expression using <>. Returns SQL condition string,
// arguments array, and error if any.
func fromNe(column string, value any, tableAlias string) (string, []any, error) {
	return "? <> ?", []any{qualifyColumn(mapColumnFromLegacyMongo(column), tableAlias), value}, nil
}
