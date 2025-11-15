package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStore_UserInvitationsUpsert(t *testing.T) {
	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock

	now := time.Now()
	mockClock.On("Now").Return(now)

	cases := []struct {
		description string
		email       string
		fixtures    []string
		expected    map[string]any
	}{
		{
			description: "succeeds creating new invitation",
			email:       "john.doe@test.com",
			fixtures:    []string{},
			expected: map[string]any{
				"email":       "john.doe@test.com",
				"created_at":  primitive.NewDateTimeFromTime(now),
				"updated_at":  primitive.NewDateTimeFromTime(now),
				"invitations": int32(1),
			},
		},
		{
			description: "succeeds updating existing invitation",
			email:       "jane.doe@test.com",
			fixtures:    []string{fixtureUserInvitations},
			expected: map[string]any{
				"email":       "jane.doe@test.com",
				"updated_at":  primitive.NewDateTimeFromTime(now),
				"invitations": float64(3),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			upsertedID, err := s.UserInvitationsUpsert(ctx, tc.email)
			require.NoError(tt, err)
			require.NotEmpty(tt, upsertedID)

			objID, _ := primitive.ObjectIDFromHex(upsertedID)

			tmpInvitation := make(map[string]any)
			require.NoError(tt, db.Collection("user_invitations").FindOne(ctx, bson.M{"_id": objID}).Decode(&tmpInvitation))

			require.Equal(tt, objID, tmpInvitation["_id"])
			for field, expectedValue := range tc.expected {
				require.Equal(tt, expectedValue, tmpInvitation[field])
			}

			if tc.description == "succeeds updating existing invitation" {
				require.NotEqual(tt, primitive.NewDateTimeFromTime(now), tmpInvitation["created_at"])
			}
		})
	}
}
