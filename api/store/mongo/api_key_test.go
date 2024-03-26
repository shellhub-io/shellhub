package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestAPIKeyCreate(t *testing.T) {
	cases := []struct {
		description string
		APIKey      *models.APIKey
		expected    error
	}{
		{
			description: "success when try create a APIKey",
			APIKey: &models.APIKey{
				UserID: "id",
				Name:   "APIKeyName",
			},
			expected: nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("api_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var testData []interface{}
			doc := bson.M{"user_id": tc.APIKey.UserID, "name": tc.APIKey.Name}
			testData = append(testData, doc)

			if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}

			err := mongostore.APIKeyCreate(ctx, tc.APIKey)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestAPIKeyList(t *testing.T) {
	cases := []struct {
		description   string
		requestParams *requests.APIKeyList
		expected      error
	}{
		{
			description: "failure when  ID is invalid",
			requestParams: &requests.APIKeyList{
				UserID:      "",
				TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
				Paginator:   query.Paginator{Page: 1, PerPage: 10},
				Sorter:      query.Sorter{By: "expires_in", Order: query.OrderAsc},
			},
			expected: nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("api_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.requestParams.UserID != "" {
				var testData []interface{}
				doc := bson.M{
					"user_id":   tc.requestParams.UserID,
					"name":      tc.requestParams.Tenant,
					"paginator": tc.requestParams.Paginator,
					"sorter":    tc.requestParams.Sorter,
				}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}
			_, _, err := mongostore.APIKeyList(ctx, tc.requestParams.UserID, tc.requestParams.Paginator, tc.requestParams.Sorter, "tenant")
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	cases := []struct {
		description string
		id          string
		expected    error
	}{
		{
			description: "fails when try delete with a invalid id",
			id:          "",
			expected:    store.ErrNoDocuments,
		},
	}

	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("api_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.id != "" {
				var testData []interface{}
				doc := bson.M{"id": tc.id}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(context.TODO(), collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}
			err := mongostore.APIKeyDelete(context.TODO(), tc.id, "tenant")
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestRenameAPIKey(t *testing.T) {
	cases := []struct {
		description   string
		requestParams *requests.APIKeyChanges
		expected      error
	}{
		{
			description: "success",
			requestParams: &requests.APIKeyChanges{
				ID:   "507f1f77bcf86cd7994390bb",
				Name: "rename",
			},
			expected: nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()

	collection := mongostore.db.Collection("api_keys")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			if tc.requestParams.ID != "" {
				var testData []interface{}
				doc := bson.M{"_id": tc.requestParams.ID, "name": "oldname"}
				testData = append(testData, doc)

				if err := dbtest.InsertMockData(ctx, collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.APIKeyEdit(ctx, tc.requestParams)
			assert.Equal(t, tc.expected, err)
		})
	}
}
