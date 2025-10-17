package pg_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
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
				MaxNamespaces:  -1,
				EmailMarketing: true,
				Preferences: models.UserPreferences{
					RecoveryEmail:      "jane.smith@test.com",
					AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
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
				MaxNamespaces:  -1,
				EmailMarketing: true,
				Preferences: models.UserPreferences{
					RecoveryEmail:      "jane.smith@test.com",
					AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
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
					MaxNamespaces:  -1,
					PasswordDigest: "$2y$12$VVm2ETx7AvaGlfMYqNYK9uzU2M45YZ70YnT..O.s1o2zdE1pekhq6",
					EmailMarketing: true,
					Preferences: models.UserPreferences{
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
						RecoveryEmail:      "jane.smith@test.com",
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

func TestPg_UserSave(t *testing.T) {
	newUser := func(ctx context.Context) (*models.User, error) {
		user := &models.User{
			Status:         models.UserStatusConfirmed,
			Origin:         models.UserOriginLocal,
			Name:           "Test User For Save",
			Username:       "test_save_user",
			Email:          "test.save@test.com",
			PasswordDigest: "$2y$12$VVm2ETx7AvaGlfMYqNYK9uzU2M45YZ70YnT..O.s1o2zdE1pekhq6",
			EmailMarketing: true,
			MaxNamespaces:  -1,
			Preferences: models.UserPreferences{
				RecoveryEmail:      "recovery.email@test.com",
				AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
				PreferredNamespace: "",
			},
		}

		id, err := s.UserCreate(ctx, user)
		if err != nil {
			return nil, err
		}

		// Recupera o usuário completo do banco
		return s.UserGet(ctx, store.UserIdentID, id)
	}

	type Expected struct {
		err error
	}

	cases := []struct {
		it       string
		ctx      context.Context
		setup    func(context.Context) (*models.User, error)
		update   func(*models.User) *models.User
		mocks    func()
		ensure   func(context.Context, *models.User, *testing.T)
		expected Expected
	}{
		{
			it:  "should update user and set UpdatedAt",
			ctx: context.Background(),
			setup: func(ctx context.Context) (*models.User, error) {
				return newUser(ctx)
			},
			update: func(user *models.User) *models.User {
				// Cria uma cópia com as alterações que queremos aplicar
				updatedUser := *user
				updatedUser.Name = "Updated Name"
				updatedUser.Username = "updated_username"
				updatedUser.Email = "updated.email@test.com"
				return &updatedUser
			},
			mocks: func() {
				mockClock := new(clockmock.Clock)
				clock.DefaultBackend = mockClock
				mockClock.On("Now").Return(time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC))
			},
			ensure: func(ctx context.Context, user *models.User, t *testing.T) {
				// Recupera o usuário do banco para verificar se as alterações foram aplicadas
				updatedUser, err := s.UserGet(ctx, store.UserIdentID, user.ID)
				require.NoError(t, err)

				require.Equal(t, "Updated Name", updatedUser.Name)
				require.Equal(t, "updated_username", updatedUser.Username)
				require.Equal(t, "updated.email@test.com", updatedUser.Email)
				require.Equal(t, time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC), updatedUser.UpdatedAt)
			},
			expected: Expected{
				err: nil,
			},
		},
		{
			it:  "should return error when user does not exist",
			ctx: context.Background(),
			setup: func(ctx context.Context) (*models.User, error) {
				// Cria um usuário que não existe no banco
				return &models.User{
					ID: uuid.Generate(), // ID aleatório que não deve existir
				}, nil
			},
			update: func(user *models.User) *models.User {
				// Não precisamos alterar nada, apenas retornar o mesmo usuário
				return user
			},
			mocks:  nil,
			ensure: nil,
			expected: Expected{
				err: store.ErrNoDocuments,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.it, func(tt *testing.T) {
			tt.Parallel()

			// Configuração de mocks, se necessário
			if tc.mocks != nil {
				tc.mocks()
			}

			// Preparação: cria o usuário para o teste
			originalUser, err := tc.setup(tc.ctx)
			require.NoError(tt, err)

			// Atualiza o usuário com as modificações definidas no caso de teste
			userToUpdate := tc.update(originalUser)

			require.Equal(tt, tc.expected.err, s.UserSave(tc.ctx, userToUpdate))
			tc.ensure(tc.ctx, userToUpdate, tt)
		})
	}
}

func TestPg_UserDelete(t *testing.T) {
	newUser := func(ctx context.Context) (string, error) {
		user := &models.User{
			ID:             uuid.Generate(),
			CreatedAt:      clock.Now(),
			UpdatedAt:      clock.Now(),
			LastLogin:      clock.Now(),
			Status:         models.UserStatusConfirmed,
			Origin:         models.UserOriginLocal,
			ExternalID:     "",
			Name:           "Test User For Delete",
			Username:       "test_delete_user",
			Email:          "test.delete@test.com",
			PasswordDigest: "$2y$12$VVm2ETx7AvaGlfMYqNYK9uzU2M45YZ70YnT..O.s1o2zdE1pekhq6",
			EmailMarketing: true,
			MaxNamespaces:  -1,
			Preferences: models.UserPreferences{
				RecoveryEmail:      "recovery.email@test.com",
				AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
				PreferredNamespace: "",
			},
		}

		id, err := s.UserCreate(ctx, user)
		return id, err
	}

	type Expected struct {
		err    error
		ensure func(*testing.T, string)
	}

	cases := []struct {
		it       string
		ctx      context.Context
		setup    func(context.Context) (*models.User, error)
		expected Expected
	}{
		{
			it:  "should return error when user does not exist",
			ctx: context.Background(),
			setup: func(ctx context.Context) (*models.User, error) {
				return &models.User{ID: uuid.Generate()}, nil
			},
			expected: Expected{
				err:    store.ErrNoDocuments,
				ensure: func(*testing.T, string) {},
			},
		},
		{
			it:  "should delete user when exists",
			ctx: context.Background(),
			setup: func(ctx context.Context) (*models.User, error) {
				id, err := newUser(ctx)
				if err != nil {
					return nil, err
				}

				user, err := s.UserGet(ctx, store.UserIdentID, id)
				return user, err
			},
			expected: Expected{
				err: nil,
				ensure: func(t *testing.T, id string) {
					exists, err := driver.NewSelect().Model((*entity.User)(nil)).Where("id = ?", id).Exists(context.TODO())
					require.NoError(t, err)
					require.False(t, exists)
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.it, func(tt *testing.T) {
			tt.Parallel()

			user, err := tc.setup(tc.ctx)
			require.NoError(tt, err)

			require.Equal(tt, tc.expected.err, s.UserDelete(tc.ctx, user))
			tc.expected.ensure(tt, user.ID)
		})
	}
}
