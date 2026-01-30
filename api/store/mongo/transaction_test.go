package mongo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestWithTransaction(t *testing.T) {
	ctx := context.Background()

	// Create MongoDB container
	srv := &dbtest.Server{}
	srv.Container.Database = "test"

	if err := srv.Up(ctx); err != nil {
		t.Fatalf("Failed to start MongoDB container: %v", err)
	}
	defer srv.Down(ctx)

	// Create store
	s, err := mongo.NewStore(ctx, srv.Container.ConnectionString+"/"+srv.Container.Database, cache.NewNullCache())
	if err != nil {
		t.Fatalf("Failed to create MongoDB store: %v", err)
	}

	// Get database handle
	store := s.(*mongo.Store)
	db := store.GetDB()

	cases := []struct {
		description string
		callback    func(ctx context.Context) error
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
				require.Equal(tt, tc.expected.Error(), err.Error())
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
