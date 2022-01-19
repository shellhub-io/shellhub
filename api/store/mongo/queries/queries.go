package queries

import (
	"errors"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrFilterInvalid         = errors.New("filter is invalid")
	ErrFilterPropertyInvalid = errors.New("filter property is not valid")
)

// BuildFilterQuery creates a MongoDB's filter query from models.Filter for filtering a fields in a database.
func BuildFilterQuery(filters []models.Filter) ([]bson.M, error) {
	const (
		TypeProperty = "property"
		TypeOperator = "operator"
	)

	properties := map[string]func(value interface{}) (bson.M, error){
		"contains": func(value interface{}) (bson.M, error) {
			switch value.(type) {
			case string:
				return bson.M{"$regex": value, "$options": "i"}, nil
			case []interface{}:
				return bson.M{"$all": value}, nil
			}

			return nil, ErrFilterPropertyInvalid
		},
		"eq": func(value interface{}) (bson.M, error) { //nolint:unparam
			return bson.M{"$eq": value}, nil
		},
		"bool": func(value interface{}) (bson.M, error) {
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
		},
		"gt": func(value interface{}) (bson.M, error) {
			switch v := value.(type) {
			case int:
				value = v
			case string:
				var err error
				value, err = strconv.Atoi(v)
				if err != nil {
					return nil, err
				}
			}

			return bson.M{"$gt": value}, nil
		},
	}

	operations := map[string]func() (string, error){
		"and": func() (string, error) {
			return "$and", nil
		},
		"or": func() (string, error) {
			return "$or", nil
		},
	}

	var queryFilter []bson.M
	var queryMatcher []bson.M

	for _, filter := range filters {
		switch filter.Type {
		case TypeProperty:
			// Converts a filter's param type to PropertyParams.
			params, ok := filter.Params.(*models.PropertyParams)
			if !ok {
				return nil, ErrFilterInvalid
			}

			// Trys to get a function that returns the query through operator.
			fn := properties[params.Operator]
			if fn == nil {
				// If the operator is not found, jump to next iteration.
				continue
			}

			// If the property is valid, get the data returned from function to use at the query.
			property, err := fn(params.Value)
			if err != nil {
				return nil, err
			}

			queryFilter = append(queryFilter, bson.M{params.Name: property})
		case TypeOperator:
			// Converts a filter's param type to OperatorParams.
			params, ok := filter.Params.(*models.OperatorParams)
			if !ok {
				return nil, ErrFilterInvalid
			}

			// Trys to get a function that returns the query through param's name.
			fn := operations[params.Name]
			if fn == nil {
				// If the operation's name is not found, jump to next iteration.
				continue
			}

			operation, err := fn()
			if err != nil {
				return nil, err
			}

			queryMatcher = append(queryMatcher, bson.M{
				"$match": bson.M{operation: queryFilter},
			})
			queryFilter = nil
		default:
			return nil, ErrFilterInvalid
		}
	}

	if len(queryFilter) > 0 {
		queryMatcher = []bson.M{
			{
				"$match": bson.M{"$or": queryFilter},
			},
		}
	}

	return queryMatcher, nil
}

// BuildPaginationQuery creates a MongoDB's query from a paginator.Query with pagination to limit the number of returned results.
func BuildPaginationQuery(pagination paginator.Query) []bson.M {
	if pagination.PerPage == -1 {
		return nil
	}

	return []bson.M{
		{"$skip": pagination.PerPage * (pagination.Page - 1)},
		{"$limit": pagination.PerPage},
	}
}
