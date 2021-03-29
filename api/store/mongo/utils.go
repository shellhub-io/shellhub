package mongo

import (
	"context"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// aggregateCount takes a pipeline and count the results
func aggregateCount(ctx context.Context, coll *mongo.Collection, pipeline []bson.M) (int, error) {
	resp := struct {
		Count int `bson:"count"`
	}{}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return 0, nil
	}

	if err = cursor.Decode(&resp); err != nil {
		return 0, err
	}

	return resp.Count, nil
}

// buildFilterQuery builds a query based on filters
func buildFilterQuery(filters []models.Filter) ([]bson.M, error) {
	var queryMatch []bson.M
	var queryFilter []bson.M

	for _, filter := range filters {
		switch filter.Type {
		case "property":
			var property bson.M
			params, ok := filter.Params.(*models.PropertyParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Operator {
			case "like":
				property = bson.M{"$regex": params.Value, "$options": "i"}
			case "eq":
				property = bson.M{"$eq": params.Value}
			case "bool":
				var value bool

				switch v := params.Value.(type) {
				case int:
					value = v != 0
				case string:
					var err error
					value, err = strconv.ParseBool(v)
					if err != nil {
						return nil, err
					}
				}

				property = bson.M{"$eq": value}
			case "gt":
				var value int

				switch v := params.Value.(type) {
				case int:
					value = v
				case string:
					var err error
					value, err = strconv.Atoi(v)
					if err != nil {
						return nil, err
					}
				}

				property = bson.M{"$gt": value}
			}

			queryFilter = append(queryFilter, bson.M{
				params.Name: property,
			})
		case "operator":
			var operator string
			params, ok := filter.Params.(*models.OperatorParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Name {
			case "and":
				operator = "$and"
			case "or":
				operator = "$or"
			}

			queryMatch = append(queryMatch, bson.M{
				"$match": bson.M{operator: queryFilter},
			})

			queryFilter = nil
		}
	}

	if len(queryFilter) > 0 {
		queryMatch = append(queryMatch, bson.M{
			"$match": bson.M{"$or": queryFilter},
		})
	}

	return queryMatch, nil
}

// buildPaginationQuery builds a query with pagination to limit the number of returned results
func buildPaginationQuery(pagination paginator.Query) []bson.M {
	if pagination.PerPage == -1 {
		return nil
	}

	return []bson.M{
		{"$skip": pagination.PerPage * (pagination.Page - 1)},
		{"$limit": pagination.PerPage},
	}
}

func renameField(db *mongo.Database, coll, from, to string) error {
	_, err := db.Collection(coll).UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{from: to}})
	return err
}
