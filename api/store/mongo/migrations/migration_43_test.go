package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration43(t *testing.T) {
	type FirewallRuleFields struct {
		Priority int    `json:"priority"`
		Action   string `json:"action" validate:"required,oneof=allow deny"`
		Active   bool   `json:"active"`
		SourceIP string `json:"source_ip" bson:"source_ip" validate:"required,regexp"`
		Username string `json:"username" validate:"required,regexp"`
		Hostname string `json:"hostname" validate:"required,regexp"`
	}

	type FirewallRule struct {
		ID                 string `json:"id,omitempty" bson:"_id,omitempty"`
		TenantID           string `json:"tenant_id" bson:"tenant_id"`
		FirewallRuleFields `bson:",inline"`
	}

	cases := []struct {
		description string
		Test        func(t *testing.T)
	}{
		{
			"Success to apply up on migration 43",
			func(t *testing.T) {
				t.Helper()

				ruleOld := &FirewallRule{
					ID:       "ruleID",
					TenantID: "tenant",
					FirewallRuleFields: FirewallRuleFields{
						Hostname: ".*",
					},
				}

				ruleNew := &models.FirewallRule{
					ID:       "ruleID",
					TenantID: "tenant",
					FirewallRuleFields: models.FirewallRuleFields{
						Filter: models.FirewallFilter{
							Hostname: ".*",
						},
					},
				}

				_, err := c.Database("test").Collection("firewall_rules").InsertOne(context.TODO(), ruleOld)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[42:43]...)
				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))

				rule := new(models.FirewallRule)
				result := c.Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": ruleOld.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(rule)
				assert.NoError(t, err)

				assert.Equal(t, ruleNew, rule)
			},
		},
		{
			"Success to apply down on migration 43",
			func(t *testing.T) {
				t.Helper()

				ruleOld := &FirewallRule{
					ID:       "ruleID",
					TenantID: "tenant",
					FirewallRuleFields: FirewallRuleFields{
						Hostname: ".*",
					},
				}

				ruleNew := &models.FirewallRule{
					ID:       "ruleID",
					TenantID: "tenant",
					FirewallRuleFields: models.FirewallRuleFields{
						Filter: models.FirewallFilter{
							Hostname: ".*",
						},
					},
				}

				_, err := c.Database("test").Collection("firewall_rules").InsertOne(context.TODO(), ruleOld)
				assert.NoError(t, err)

				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[42:43]...)
				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))

				rule := new(FirewallRule)
				result := c.Database("test").Collection("firewall_rules").FindOne(context.TODO(), bson.M{"tenant_id": ruleNew.TenantID})
				assert.NoError(t, result.Err())

				err = result.Decode(rule)
				assert.NoError(t, err)

				assert.Equal(t, ruleOld, rule)
			},
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})
			tc.Test(t)
		})
	}
}
