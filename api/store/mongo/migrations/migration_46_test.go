package migrations

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration46(t *testing.T) {
	type PublicKeyFilter struct {
		Hostname string   `json:"hostname,omitempty" bson:"hostname,omitempty" validate:"required_without=Tags,excluded_with=Tags,regexp"`
		Tags     []string `json:"tags,omitempty" bson:"tags,omitempty" validate:"required_without=Hostname,excluded_with=Hostname,max=3,unique,dive,min=3,max=255,alphanum,ascii,excludes=/@&:"`
	}

	type PublicKeyFields struct {
		Name     string          `json:"name"`
		Username string          `json:"username" bson:"username" validate:"regexp"`
		Filter   PublicKeyFilter `json:"filter" bson:"filter" validate:"required"`
	}

	type PublicKey struct {
		Data            []byte    `json:"data"`
		Fingerprint     string    `json:"fingerprint"`
		CreatedAt       time.Time `json:"created_at" bson:"created_at"`
		TenantID        string    `json:"tenant_id" bson:"tenant_id"`
		PublicKeyFields `bson:",inline"`
	}

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 46",
			func(t *testing.T) {
				t.Helper()

				keyUsernameEmpty := &PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: PublicKeyFields{
						Name:     "key",
						Username: "",
						Filter: PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				keyUsernameRegexp := &PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: PublicKeyFields{
						Name:     "key",
						Username: ".*",
						Filter: PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				_, err := c.Database("test").Collection("public_keys").InsertOne(context.Background(), keyUsernameEmpty)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[45:46]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				key := new(PublicKey)
				result := c.Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameEmpty.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				sort.Strings(key.Filter.Tags)

				assert.Equal(t, keyUsernameRegexp, key)
			},
		},
		{
			"Success to apply down on migration 46",
			func(t *testing.T) {
				t.Helper()

				keyUsernameEmpty := &PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: PublicKeyFields{
						Name:     "key",
						Username: "",
						Filter: PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				keyUsernameRegexp := &PublicKey{
					Fingerprint: "fingerprint",
					TenantID:    "tenant",
					PublicKeyFields: PublicKeyFields{
						Name:     "key",
						Username: ".*",
						Filter: PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				_, err := c.Database("test").Collection("public_keys").InsertOne(context.Background(), keyUsernameEmpty)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[45:46]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				key := new(PublicKey)
				result := c.Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameRegexp.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(key)
				assert.NoError(t, err)

				assert.Equal(t, keyUsernameEmpty, key)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})
			tc.Test(t)
		})
	}
}
