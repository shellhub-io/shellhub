package queries

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBuildQuery(t *testing.T) {
	type Expected struct {
		data []bson.M
		err  error
	}
	cases := []struct {
		description string
		filters     []models.Filter
		expected    Expected
	}{
		{
			description: "Fail when filter type is not valid",
			filters: []models.Filter{
				{
					Type: "invalid",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "valid",
						Value:    "test",
					},
				},
			},
			expected: Expected{nil, ErrFilterInvalid},
		},
		{
			description: "Fail when operator in property is invalid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "invalid",
						Value:    "valid",
					},
				},
			},
			expected: Expected{
				data: nil,
				err:  nil,
			},
		},
		{
			description: "Success when one operator in property is valid and other is invalid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "invalid",
						Value:    "test",
					},
				},
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "eq",
						Value:    "valid",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "valid"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Success when operator in property is valid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "eq",
						Value:    "valid",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "valid"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Fail when operator in operator is invalid",
			filters: []models.Filter{
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "invalid",
					},
				},
			},
			expected: Expected{
				data: nil,
				err:  nil,
			},
		},
		{
			description: "Fail when operator in operator is valid and other invalid",
			filters: []models.Filter{
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "and",
					},
				},
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "invalid",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M(nil)}}},
				err:  nil,
			},
		},
		{
			description: "Success when operator in operator is valid",
			filters: []models.Filter{
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "and",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M(nil)}}},
				err:  nil,
			},
		},
		{
			description: "Fail when property operator is invalid and operator is valid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "invalid",
						Value:    "test",
					},
				},
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "and",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M(nil)}}},
				err:  nil,
			},
		},
		{
			description: "Fail when property operator is valid and operator is invalid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "eq",
						Value:    "test",
					},
				},
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "invalid",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$or": []bson.M{{"test": bson.M{"$eq": "test"}}}}}},
				err:  nil,
			},
		},
		{
			description: "Success when property and operator is valid",
			filters: []models.Filter{
				{
					Type: "property",
					Params: &models.PropertyParams{
						Name:     "test",
						Operator: "eq",
						Value:    "test",
					},
				},
				{
					Type: "operator",
					Params: &models.OperatorParams{
						Name: "and",
					},
				},
			},
			expected: Expected{
				data: []bson.M{{"$match": bson.M{"$and": []bson.M{{"test": bson.M{"$eq": "test"}}}}}},
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			query, err := BuildFilterQuery(tc.filters)

			assert.Equal(t, tc.expected, Expected{query, err})
		})
	}
}
