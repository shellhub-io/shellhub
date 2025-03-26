package pg_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestPg_UserCreate(t *testing.T) {
	cases := []struct {
		it     string
		ctx    context.Context
		user   *models.User
		assert func(context.Context, *testing.T)
	}{
		{
			it:  "should generate an UUIDv7",
			ctx: context.Background(),
			user: &models.User{
				Status:         models.UserStatusConfirmed,
				Origin:         models.UserOriginLocal,
				ExternalID:     "",
				Name:           "Paul Bryson",
				Email:          "paul.bryson@test.com",
				PasswordDigest: "$2y$12$VVm2ETx7AvaGlfMYqNYK9uzU2M45YZ70YnT..O.s1o2zdE1pekhq6",
				Preferences: models.UserPreferences{
					MaxNamespaces:      -1,
					SecurityEmail:      "jane.smith@test.com",
					AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					EmailMarketing:     true,
					PreferredNamespace: "",
				},
			},
			assert: func(ctx context.Context, t *testing.T) {
				u := new(models.User)
				require.NoError(t, driver.NewSelect().Model(u).Where("email = ?", "paul.bryson@test.com").Scan(ctx))

				uuidPattern := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
				require.True(t, uuidPattern.MatchString(u.ID))
			},
		},
		{
			it:  "should set CreatedAt and UpdatedAt timestamps",
			ctx: context.Background(),
			user: &models.User{
				Status:         models.UserStatusConfirmed,
				Origin:         models.UserOriginLocal,
				Name:           "Stephen Aaron",
				Username:       "stephen_aaron",
				Email:          "stephen.aaron@test.com",
				PasswordDigest: "$2y$12$AnotherHashedPassword",
				Preferences: models.UserPreferences{
					MaxNamespaces:      -1,
					SecurityEmail:      "jane.smith@test.com",
					AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					EmailMarketing:     true,
					PreferredNamespace: "",
				},
			},
			assert: func(ctx context.Context, t *testing.T) {
				u := new(models.User)
				require.NoError(t, driver.NewSelect().Model(u).Where("email = ?", "stephen.aaron@test.com").Scan(ctx))

				require.WithinDuration(t, time.Now(), u.CreatedAt, 1*time.Second)
				require.WithinDuration(t, time.Now(), u.UpdatedAt, 1*time.Second)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.it, func(tt *testing.T) {
			tt.Parallel()

			insertedID, err := s.UserCreate(tc.ctx, tc.user)
			require.NoError(t, err)
			tc.assert(tc.ctx, tt)

			_, err = driver.NewDelete().Table("users").Where("id = ?", insertedID).Exec(tc.ctx)
			require.NoError(tt, err)
		})
	}
}

func TestPg_UserGet(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	cases := []struct {
		it       string
		ctx      context.Context
		ident    store.UserIdent
		val      string
		expected Expected
	}{
		{
			it:    "should returns an error when the user does not exists",
			ctx:   context.Background(),
			ident: store.UserIdentEmail,
			val:   "nonexistent@test.com",
			expected: Expected{
				user: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			it:    "should returns an error when the user does not exists",
			ctx:   context.Background(),
			ident: store.UserIdentEmail,
			val:   "john.doe@test.com",
			expected: Expected{
				user: &models.User{
					ID:             "0195cefa-aa01-7efb-8098-c9c173056250",
					CreatedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
					UpdatedAt:      time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
					LastLogin:      time.Time{},
					Status:         models.UserStatusConfirmed,
					Origin:         models.UserOriginLocal,
					ExternalID:     "",
					Name:           "Jonh Doe",
					Username:       "john_doe",
					Email:          "john.doe@test.com",
					PasswordDigest: "$2y$12$VVm2ETx7AvaGlfMYqNYK9uzU2M45YZ70YnT..O.s1o2zdE1pekhq6",
					Preferences: models.UserPreferences{
						MaxNamespaces:      -1,
						SecurityEmail:      "jane.smith@test.com",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
						EmailMarketing:     true,
						PreferredNamespace: "",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.it, func(tt *testing.T) {
			tt.Parallel()

			user, err := s.UserGet(tc.ctx, tc.ident, tc.val)
			require.Equal(tt, tc.expected, Expected{user, err})
		})
	}
}
