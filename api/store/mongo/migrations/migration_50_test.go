package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envMocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMigration50(t *testing.T) {
	mock := &envMocks.Backend{}
	envs.DefaultBackend = mock

	user1ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
	assert.NoError(t, err)
	user1 := &models.User{
		ID: user1ID.String(),
	}

	user2ID, err := primitive.ObjectIDFromHex("507f1f77bcf86cd799439012")
	assert.NoError(t, err)
	user2 := &models.User{
		ID: user2ID.String(),
	}

	namespace1 := &models.Namespace{
		Name:  "namespace1",
		Owner: user1ID.String(),
		Billing: &models.Billing{
			Active: true,
		},
	}
	namespace2 := &models.Namespace{
		Name:  "namespace2",
		Owner: user1ID.String(),
	}
	namespace3 := &models.Namespace{
		Name:  "namespace3",
		Owner: user2ID.String(),
	}

	cases := []struct {
		description string
		before      func()
		test        func() (int, error)
		expected    int
		after       func()
	}{
		{
			"Success to apply up on migration 50 when it is a ShellHub Cloud instance",
			func() {
				_, err := srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
				assert.NoError(t, err)
			},
			func() (int, error) {
				mock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				migrations := GenerateMigrations()[49:50]
				migrates := migrate.NewMigrate(srv.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return 0, err
				}

				user := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				if err != nil {
					return 0, err
				}

				err = result.Decode(user)
				if err != nil {
					return 0, err
				}

				return user.MaxNamespaces, nil
			},
			2,
			func() {
				err = srv.Client().Database("test").Collection("users").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("namespaces").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("migrations").Drop(context.TODO())
				assert.NoError(t, err)
			},
		},
		{
			"Success to apply up on migration 50 when it is a ShellHub Community instance",
			func() {
				_, err := srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
				assert.NoError(t, err)
			},
			func() (int, error) {
				mock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()

				migrations := GenerateMigrations()[49:50]
				migrates := migrate.NewMigrate(srv.Client().Database("test"), migrations...)
				err := migrates.Up(context.Background(), migrate.AllAvailable)
				if err != nil {
					return 0, err
				}

				user := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				if err != nil {
					return 0, err
				}

				err = result.Decode(user)
				if err != nil {
					return 0, err
				}

				return user.MaxNamespaces, nil
			},
			-1,
			func() {
				err = srv.Client().Database("test").Collection("users").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("namespaces").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("migrations").Drop(context.TODO())
				assert.NoError(t, err)
			},
		},
		{
			"Success to apply down on migration 50",
			func() {
				_, err := srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("users").InsertOne(context.TODO(), user2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace1)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace2)
				assert.NoError(t, err)
				_, err = srv.Client().Database("test").Collection("namespaces").InsertOne(context.TODO(), namespace3)
				assert.NoError(t, err)
			},
			func() (int, error) {
				migrations := GenerateMigrations()[49:50]
				migrates := migrate.NewMigrate(srv.Client().Database("test"), migrations...)
				err := migrates.Down(context.Background(), migrate.AllAvailable)
				if err != nil {
					return 0, err
				}

				user := new(models.User)
				result := srv.Client().Database("test").Collection("users").FindOne(context.TODO(), bson.M{"_id": user1ID.String()})
				if err != nil {
					return 0, err
				}

				err = result.Decode(user)
				if err != nil {
					return 0, err
				}

				return user.MaxNamespaces, nil
			},
			0,
			func() {
				err = srv.Client().Database("test").Collection("users").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("namespaces").Drop(context.TODO())
				assert.NoError(t, err)
				err = srv.Client().Database("test").Collection("migrations").Drop(context.TODO())
				assert.NoError(t, err)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, fixtures.Teardown())
			})

			tc.before()

			actual, err := tc.test()
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)

			tc.after()
		})
	}
}
