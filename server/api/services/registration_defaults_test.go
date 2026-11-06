package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_RegisterUserAccountDefaults(t *testing.T) {
	for _, invited := range []bool{false, true} {
		name := "open registration"
		if invited {
			name = "invited registration"
		}
		t.Run(name, func(t *testing.T) {
			override(t, &openSignupEnabled, true)
			override(t, &userRegisteredHooks, nil)
			hasher := hashmock.NewMockHasher(t)
			hash.Set(t, hasher)
			hasher.On("Do", "secret123").Return("hashed-password", nil).Once()
			st := storemock.NewMockStore(t)
			ctx := context.Background()
			email := "request@example.com"
			id := ""
			lookup := st.On("UserInvitationGet", ctx, store.UserInvitationEmailResolver, email)
			if invited {
				st.On("SystemGet", ctx).Return(localSignInOn(), nil).Once()
				invitation := &models.UserInvitation{ID: "invitee", Email: "Invite@Example.com", Status: models.UserInvitationStatusPending}
				lookup.Return(invitation, nil).Once()
				st.On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(func(txCtx context.Context, cb store.TransactionCb) error { return cb(txCtx) }).Once()
				st.On("UserInvitationUpdate", ctx, invitation).Return(nil).Once()
				email, id = "invite@example.com", "invitee"
			} else {
				lookup.Return(nil, store.ErrNoDocuments).Once()
			}
			st.On("UserCreate", ctx, mock.AnythingOfType("*models.User")).
				Run(func(args mock.Arguments) {
					user, ok := args.Get(1).(*models.User)
					require.True(t, ok)
					assert.Equal(t, id, user.ID)
					assert.Equal(t, models.UserData{Name: "Alice", Email: email, Username: "alice"}, user.UserData)
					assert.Equal(t, models.UserPassword{Plain: "secret123", Hash: "hashed-password"}, user.Password)
					assert.Equal(t, now, user.CreatedAt)
					assert.Equal(t, 1, user.MaxNamespaces)
					assert.True(t, user.EmailMarketing)
					assert.False(t, user.Admin)
					assert.False(t, user.AwaitingApproval)
					assert.Equal(t, models.UserOriginLocal, user.Origin)
					assert.Equal(t, models.UserStatusNotConfirmed, user.Status)
					assert.Equal(t, models.UserPreferences{AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal}}, user.Preferences)
				}).Return("created-user", nil).Once()
			s := NewService(st, privateKey, publicKey, storecache.NewNullCache(), WithClock(clock.NewFixed(now)))
			_, err := s.RegisterUser(ctx, requests.RegisterUser{
				Name: "Alice", Username: "ALICE", Email: "Request@Example.com", Password: "secret123", EmailMarketing: true,
			})
			require.NoError(t, err)
		})
	}
}
