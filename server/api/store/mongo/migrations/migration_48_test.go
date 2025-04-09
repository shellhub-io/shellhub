package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration48(t *testing.T) {
	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 48",
			func(t *testing.T) {
				t.Helper()

				namespace := models.Namespace{
					TenantID: "tenant",
				}

				rule0 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 0,
					},
				}

				rule1 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 1,
					},
				}

				rule2 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 2,
					},
				}

				ctx := context.Background()

				_, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule0)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule1)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule2)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[47:48]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				key := new(models.FirewallRule)
				result := c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"tenant_id": namespace.TenantID})
				assert.NoError(t, result.Err())

				assert.NoError(t, result.Decode(key))
				assert.Equal(t, 2, key.Priority)
			},
		},
		{
			"Success to apply down on migration 48",
			func(t *testing.T) {
				t.Helper()

				namespace := models.Namespace{
					TenantID: "tenant",
				}

				rule0 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 0,
					},
				}

				rule1 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 1,
					},
				}

				rule2 := models.FirewallRule{
					TenantID: namespace.TenantID,
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 2,
					},
				}

				ctx := context.Background()

				_, err := c.Database("test").Collection("namespaces").InsertOne(ctx, namespace)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule0)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule1)
				assert.NoError(t, err)
				_, err = c.Database("test").Collection("firewall_rules").InsertOne(ctx, rule2)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[47:48]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				key := new(models.FirewallRule)
				result := c.Database("test").Collection("firewall_rules").FindOne(ctx, bson.M{"tenant_id": namespace.TenantID})
				assert.NoError(t, result.Err())

				assert.NoError(t, result.Decode(key))
				assert.Equal(t, 0, key.Priority)
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
