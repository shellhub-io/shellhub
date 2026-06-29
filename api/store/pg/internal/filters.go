package internal

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
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

// mapFieldToColumn translates a filter field name to the corresponding PostgreSQL column name.
// It falls back to legacyMongoFieldMapping for Mongo-compatible field paths.
// Returns the original field name if no mapping exists.
//
// NOTE: the device_uid→device_id alias is NOT applied here because it is
// specific to the sessions table.  It is handled in ParseFilterProperty when
// tableAlias == "session" so it cannot accidentally affect other filter contexts.
func mapFieldToColumn(field string) string {
	if mapped, ok := legacyMongoFieldMapping[field]; ok {
		return mapped
	}

	return field
}

// TODO: remove when MongoDB support is dropped.
// "online" is a virtual field (not a real column), so we can't filter by it directly in WHERE.
// We expand it to the actual expression: disconnected_at IS NULL AND last_seen > (now - 2min).
func fromOnlineFilter(value any) (string, []any, bool, error) {
	var isOnline bool

	switch v := value.(type) {
	case bool:
		isOnline = v
	case float64:
		// JSON numbers always decode to float64; nonzero means online.
		isOnline = v != 0
	case string:
		var err error
		isOnline, err = strconv.ParseBool(v)
		if err != nil {
			return "", nil, false, err
		}
	default:
		return "", nil, false, ErrUnsupportedBoolType
	}

	threshold := clock.Now().Add(-2 * time.Minute)

	if isOnline {
		return `("device"."disconnected_at" IS NULL AND "device"."last_seen" > ?)`, []any{threshold}, true, nil
	}

	return `("device"."disconnected_at" IS NOT NULL OR "device"."last_seen" <= ?)`, []any{threshold}, true, nil
}

// fromActiveFilter returns an EXISTS or NOT EXISTS correlated subquery on active_sessions.
//
// When tableAlias is "session" (session-list context) the subquery correlates on the
// current session row's id, because the outer query is already iterating sessions:
//
//	EXISTS  (SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")
//	NOT EXISTS ...
//
// In all other contexts (e.g. device-list) the subquery checks whether any session for
// the current device row has an active_sessions entry:
//
//	EXISTS  (SELECT 1 FROM "active_sessions"
//	         JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id"
//	         WHERE "sessions"."device_id" = "device"."id")
func fromActiveFilter(value any, tableAlias string) (string, []any, bool, error) {
	var isActive bool

	switch v := value.(type) {
	case bool:
		isActive = v
	case float64:
		isActive = v != 0
	case string:
		var err error
		isActive, err = strconv.ParseBool(v)
		if err != nil {
			return "", nil, false, err
		}
	default:
		return "", nil, false, ErrUnsupportedBoolType
	}

	if tableAlias == "session" {
		// Session-list context: correlate directly on the session's own id.
		const sub = `(SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`

		if isActive {
			return `EXISTS ` + sub, nil, true, nil
		}

		return `NOT EXISTS ` + sub, nil, true, nil
	}

	// Device-list (and other) contexts: correlate via the sessions join.
	const subquery = `(SELECT 1 FROM "active_sessions" JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id" WHERE "sessions"."device_id" = "device"."id")`

	if isActive {
		return `EXISTS ` + subquery, nil, true, nil
	}

	return `NOT EXISTS ` + subquery, nil, true, nil
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

	// active is a virtual field backed by the active_sessions table (see fromActiveFilter for details)
	if fp.Name == "active" {
		return fromActiveFilter(fp.Value, tableAlias)
	}

	// In the session context "device_uid" is a user-facing alias for the actual
	// "device_id" column in the sessions table.  The mapping is applied here
	// (not in mapFieldToColumn) so it is scoped to sessions only and cannot
	// silently affect other filter contexts that happen to expose a device_uid
	// field but use a different column name or no column at all.
	if tableAlias == "session" && fp.Name == "device_uid" {
		adjusted := *fp
		adjusted.Name = "device_id"
		fp = &adjusted
	}

	// tags.name requires an EXISTS subquery through the device_tags junction table,
	// because tags live in a separate table with a many-to-many relationship.
	if fp.Name == "tags.name" {
		return fromTagsFilter(fp.Operator, fp.Value)
	}

	// custom_fields is a JSONB column; search across all values.
	if fp.Name == "custom_fields" {
		return fromCustomFieldsFilter(fp.Operator, fp.Value)
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
	case "lt":
		condition, args, err = fromLt(fp.Name, fp.Value, tableAlias)
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
	column = mapFieldToColumn(column)

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
	return "? = ?", []any{qualifyColumn(mapFieldToColumn(column), tableAlias), value}, nil
}

// fromBool converts a "bool" JSON expression to an SQL expression. It handles various input types (int, float64,
// string, bool) and converts them to boolean values.
//
// - For integers or float64: 0 is false, anything else is true
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
	case float64:
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

	return "? = ?", []any{qualifyColumn(mapFieldToColumn(column), tableAlias), boolValue}, nil
}

// fromGt converts a "gt" (greater than) JSON expression to an SQL expression using >. It handles various numeric types
// (int, float, etc.) and string representations of numbers. For strings, it attempts to convert to int first, then to
// float if int conversion fails. Returns SQL condition string, arguments array, and error if any.
func fromGt(column string, value any, tableAlias string) (string, []any, error) {
	column = mapFieldToColumn(column)

	switch v := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return "? > ?", []any{qualifyColumn(column, tableAlias), v}, nil
	case time.Time:
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

// fromLt converts a "lt" (less than) JSON expression to an SQL expression using <. It handles numeric types,
// strings, and time.Time values. Returns SQL condition string, arguments array, and error if any.
func fromLt(column string, value any, tableAlias string) (string, []any, error) {
	column = mapFieldToColumn(column)

	switch v := value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, float32, float64:
		return "? < ?", []any{qualifyColumn(column, tableAlias), v}, nil
	case time.Time:
		return "? < ?", []any{qualifyColumn(column, tableAlias), v}, nil
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

		return "? < ?", []any{qualifyColumn(column, tableAlias), num}, nil
	default:
		return "", nil, ErrUnsupportedNumericType
	}
}

// fromNe converts a "ne" (not equals) JSON expression to an SQL expression using <>. Returns SQL condition string,
// arguments array, and error if any.
func fromNe(column string, value any, tableAlias string) (string, []any, error) {
	return "? <> ?", []any{qualifyColumn(mapFieldToColumn(column), tableAlias), value}, nil
}

// fromCustomFieldsFilter searches across all values of the custom_fields JSONB column.
// Only "contains" is supported: it matches any value using ILIKE.
func fromCustomFieldsFilter(operator string, value any) (string, []any, bool, error) {
	if operator != "contains" {
		return "", nil, false, nil
	}

	v, ok := value.(string)
	if !ok {
		return "", nil, false, ErrUnsupportedContainsType
	}

	const sql = `EXISTS (SELECT 1 FROM jsonb_each_text("device"."custom_fields") WHERE value ILIKE ?)`

	return sql, []any{"%" + v + "%"}, true, nil
}
