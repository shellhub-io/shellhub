package internal

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"go.mongodb.org/mongo-driver/bson"
)

// ParseFilterOperator constructs the filter operator, returning its Bson representation and a boolean
// indicating whether the operator is valid or not.
func ParseFilterOperator(fo *query.FilterOperator) (string, bool) {
	validProperties := []string{"and", "or"}
	for _, op := range validProperties {
		if op == fo.Name {
			return fmt.Sprintf("$%s", fo.Name), true
		}
	}

	return "", false
}

// ParseFilterProperty constructs the property, returning the BSON representation of the property, a boolean
// indicating whether the operator is valid or not, and an error if any.
func ParseFilterProperty(fp *query.FilterProperty) (bson.M, bool, error) {
	var res bson.M
	var err error
	var ok bool

	switch fp.Operator {
	case "contains":
		res, err = fromContains(fp.Value)
		ok = true
	case "eq":
		res, err = fromEq(fp.Value)
		ok = true
	case "bool":
		res, err = fromBool(fp.Value)
		ok = true
	case "gt":
		res, err = fromGt(fp.Value)
		ok = true
	case "lt":
		res, err = fromLt(fp.Value)
		ok = true
	case "ne":
		res, err = fromNe(fp.Value)
		ok = true
	default:
		return nil, false, nil
	}

	return res, ok, err
}

// fromContains converts a "contains" JSON expression to a Bson expression using "$regex" or "$all".
func fromContains(value interface{}) (bson.M, error) {
	switch value.(type) {
	case string:
		return bson.M{"$regex": value, "$options": "i"}, nil
	case []interface{}:
		return bson.M{"$all": value}, nil
	}

	return nil, errors.New("invalid value type for fromContains")
}

// fromEq converts an "eq" JSON expression to a Bson expression using "$eq".
func fromEq(value interface{}) (bson.M, error) {
	return bson.M{"$eq": value}, nil
}

// fromBool converts a "bool" JSON expression to a Bson expression using "$eq" for comparing boolean values.
func fromBool(value interface{}) (bson.M, error) {
	switch v := value.(type) {
	case int:
		value = v != 0
	case string:
		var err error
		value, err = strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
	}

	return bson.M{"$eq": value}, nil
}

// fromGt converts a "gt" JSON expression to a Bson expression using "$gt".
func fromGt(value interface{}) (bson.M, error) {
	switch v := value.(type) {
	case int:
		return bson.M{"$gt": v}, nil
	case int64:
		return bson.M{"$gt": v}, nil
	case float64:
		return bson.M{"$gt": v}, nil
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return bson.M{"$gt": intVal}, nil
		}

		if timeVal, err := time.Parse(time.RFC3339, v); err == nil {
			return bson.M{"$gt": timeVal}, nil
		}

		return bson.M{"$gt": v}, nil
	case time.Time:
		return bson.M{"$gt": v}, nil
	default:
		return bson.M{"$gt": value}, nil
	}
}

// fromLt converts a "lt" JSON expression to a Bson expression using "$lt".
func fromLt(value interface{}) (bson.M, error) {
	switch v := value.(type) {
	case int:
		return bson.M{"$lt": v}, nil
	case int64:
		return bson.M{"$lt": v}, nil
	case float64:
		return bson.M{"$lt": v}, nil
	case string:
		if intVal, err := strconv.Atoi(v); err == nil {
			return bson.M{"$lt": intVal}, nil
		}

		if timeVal, err := time.Parse(time.RFC3339, v); err == nil {
			return bson.M{"$lt": timeVal}, nil
		}

		return bson.M{"$lt": v}, nil
	case time.Time:
		return bson.M{"$lt": v}, nil
	default:
		return bson.M{"$lt": value}, nil
	}
}

func fromNe(value interface{}) (bson.M, error) {
	return bson.M{"$ne": value}, nil
}
