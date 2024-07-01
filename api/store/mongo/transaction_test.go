package mongo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestWithTransaction(t *testing.T) {
	cases := []struct {
		description string
		callback    store.TransactionCb
		expected    error
	}{
		{
			description: "should abort changes",
			callback: func(ctx context.Context) error {
				if _, err := db.Collection("users").InsertOne(ctx, bson.M{"_id": 1, "name": "John Doe"}); err != nil {
					return err
				}

				return errors.New("error")
			},
			expected: errors.New("error"),
		},
		{
			description: "should commit changes",
			callback: func(ctx context.Context) error {
				if _, err := db.Collection("users").InsertOne(ctx, bson.M{"_id": 1, "name": "John Doe"}); err != nil {
					return err
				}

				return nil
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			if err := s.WithTransaction(ctx, tc.callback); err != nil {
				require.Equal(tt, err, tc.expected)
				target := make(map[string]interface{})
				require.Error(tt, db.Collection("users").FindOne(ctx, bson.M{"_id": 1}).Decode(&target))
				_, ok := target["name"]
				require.Equal(tt, false, ok)
			} else {
				target := make(map[string]interface{})
				require.NoError(tt, db.Collection("users").FindOne(ctx, bson.M{"_id": 1}).Decode(&target))
				require.Equal(tt, "John Doe", target["name"])
			}
		})
	}
}
