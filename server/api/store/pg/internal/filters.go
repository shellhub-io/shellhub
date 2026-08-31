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

var legacyDeviceFieldAliases = map[string]string{
	"info.platform": "platform",
	"identity.mac":  "mac",
}

func renameFilterField(fp *query.FilterProperty, name string) *query.FilterProperty {
	adjusted := *fp
	adjusted.Name = name

	return &adjusted
}

func qualifyColumn(column, tableAlias string) bun.Ident {
	if tableAlias != "" {
		return bun.Ident(tableAlias + "." + column)
	}

	return bun.Ident(column)
}

func fromOnlineFilter(value any) (string, []any, bool, error) {
	var isOnline bool

	switch v := value.(type) {
	case bool:
		isOnline = v
	case float64:
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
		const sub = `(SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`

		if isActive {
			return `EXISTS ` + sub, nil, true, nil
		}

		return `NOT EXISTS ` + sub, nil, true, nil
	}

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
	if fp.Name == "online" {
		return fromOnlineFilter(fp.Value)
	}

	if fp.Name == "active" {
		return fromActiveFilter(fp.Value, tableAlias)
	}

	if tableAlias == "session" && fp.Name == "device_uid" {
		fp = renameFilterField(fp, "device_id")
	}

	if column, ok := legacyDeviceFieldAliases[fp.Name]; ok {
		fp = renameFilterField(fp, column)
	}

	if fp.Name == "tags.name" {
		return fromTagsFilter(fp.Operator, fp.Value)
	}

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

func fromContains(column string, value any, tableAlias string) (string, []any, error) {
	switch v := value.(type) {
	case string:
		return "? ILIKE ?", []any{qualifyColumn(column, tableAlias), "%" + v + "%"}, nil
	case []any:
		return "? @> ?", []any{qualifyColumn(column, tableAlias), v}, nil
	}

	return "", nil, ErrUnsupportedContainsType
}

func fromEq(column string, value any, tableAlias string) (string, []any, error) {
	return "? = ?", []any{qualifyColumn(column, tableAlias), value}, nil
}

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

	return "? = ?", []any{qualifyColumn(column, tableAlias), boolValue}, nil
}

func fromGt(column string, value any, tableAlias string) (string, []any, error) {
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

func fromLt(column string, value any, tableAlias string) (string, []any, error) {
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

func fromNe(column string, value any, tableAlias string) (string, []any, error) {
	return "? <> ?", []any{qualifyColumn(column, tableAlias), value}, nil
}

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
