package mongo

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestFirewallRuleList(t *testing.T) {
	type Expected struct {
		rules []models.FirewallRule
		len   int
		err   error
	}

	cases := []struct {
		description string
		paginator   query.Paginator
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no firewall rules are found",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			fixtures:    []string{},
			expected: Expected{
				rules: []models.FirewallRule{},
				len:   0,
				err:   nil,
			},
		},
		{
			description: "succeeds when a firewall rule is found",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rules: []models.FirewallRule{
					{
						ID:       "6504b7bd9b6c4a63a9ccc053",
						TenantID: "00000000-0000-4000-0000-000000000000",
						FirewallRuleFields: models.FirewallRuleFields{
							Priority: 1,
							Action:   "allow",
							Active:   true,
							SourceIP: ".*",
							Username: ".*",
							Filter: models.FirewallFilter{
								Hostname: "",
								Tags:     []string{"tag-1"},
							},
						},
					},
					{
						ID:       "e92f4a5d3e1a4f7b8b2b6e9a",
						TenantID: "00000000-0000-4000-0000-000000000000",
						FirewallRuleFields: models.FirewallRuleFields{
							Priority: 2,
							Action:   "allow",
							Active:   true,
							SourceIP: "192.168.1.10",
							Username: "john.doe",
							Filter: models.FirewallFilter{
								Hostname: "",
								Tags:     []string{"tag-1"},
							},
						},
					},
					{
						ID:       "78c96f0a2e5b4dca8d78f00c",
						TenantID: "00000000-0000-4000-0000-000000000000",
						FirewallRuleFields: models.FirewallRuleFields{
							Priority: 3,
							Action:   "allow",
							Active:   true,
							SourceIP: "10.0.0.0/24",
							Username: "admin",
							Filter: models.FirewallFilter{
								Hostname: "",
								Tags:     nil,
							},
						},
					},
					{
						ID:       "3fd759a1ecb64ec5a07c8c0",
						TenantID: "00000000-0000-4000-0000-000000000000",
						FirewallRuleFields: models.FirewallRuleFields{
							Priority: 4,
							Action:   "deny",
							Active:   true,
							SourceIP: "172.16.0.0/16",
							Username: ".*",
							Filter: models.FirewallFilter{
								Hostname: "",
								Tags:     []string{"tag-1"},
							},
						},
					},
				},
				len: 4,
				err: nil,
			},
		},
		{
			description: "succeeds when firewall rule list is  empty and paginator is different than -1",
			paginator:   query.Paginator{Page: 2, PerPage: 2},
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rules: []models.FirewallRule{},
				len:   0,
				err:   nil,
			},
		},
	}

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(fr []models.FirewallRule) {
		sort.Slice(fr, func(i, j int) bool {
			return fr[i].ID < fr[j].ID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			var testData []interface{}

			for _, item := range tc.expected.rules {
				testData = append(testData, item)
			}

			if len(testData) > 0 {
				if err := dbtest.InsertMockData(context.TODO(), collection, testData); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			rules, count, err := mongostore.FirewallRuleList(context.TODO(), tc.paginator)
			sort(tc.expected.rules)
			sort(rules)
			assert.Equal(t, tc.expected, Expected{rules: rules, len: count, err: err})

			if err := dbtest.DeleteMockData(context.TODO(), collection); err != nil {
				t.Fatalf("failed to insert documents: %v", err)
			}
		})
	}
}

func TestFirewallRuleGet(t *testing.T) {
	type Expected struct {
		rule *models.FirewallRule
		err  error
	}
	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rule: nil,
				err:  FromMongoError(nil),
			},
		},
		{
			description: "succeeds when firewall rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rule: &models.FirewallRule{
					ID:       "6504b7bd9b6c4a63a9ccc053",
					TenantID: "00000000-0000-4000-0000-000000000000",
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 1,
						Action:   "allow",
						Active:   true,
						SourceIP: ".*",
						Username: ".*",
						Filter: models.FirewallFilter{
							Hostname: "",
							Tags:     []string{"tag-1"},
						},
					},
				},
				err: nil,
			},
		},
	}
	ctx := context.TODO()
	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.id != "" {
				objID, err := primitive.ObjectIDFromHex(tc.id)
				if err != nil {
					t.Fatalf("failed to convert ID to ObjectID: %v", err)
				}

				doc := bson.M{
					"_id":       objID,
					"tenant_id": "00000000-0000-4000-0000-000000000000",
					"priority":  tc.expected.rule.Priority,
					"action":    tc.expected.rule.Action,
					"active":    tc.expected.rule.Active,
					"source_ip": tc.expected.rule.SourceIP,
					"username":  tc.expected.rule.Username,
					"filter": bson.M{
						"hostname": tc.expected.rule.Filter.Hostname,
						"tags":     tc.expected.rule.Filter.Tags,
					},
				}
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			rule, err := mongostore.FirewallRuleGet(context.TODO(), tc.id)
			if tc.id == "" {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, tc.expected, Expected{rule: rule, err: err})
			}
		})
	}
}

func TestFirewallRuleUpdate(t *testing.T) {
	type Expected struct {
		rule *models.FirewallRule
		err  error
	}

	cases := []struct {
		description string
		id          string
		rule        models.FirewallRuleUpdate
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when firewall rule is not found",
			id:          "",
			rule:        models.FirewallRuleUpdate{},
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rule: nil,
				err:  FromMongoError(nil),
			},
		},
		{
			description: "succeeds when firewall rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			rule: models.FirewallRuleUpdate{
				FirewallRuleFields: models.FirewallRuleFields{
					Priority: 1,
					Action:   "deny",
					Active:   true,
					SourceIP: ".*",
					Username: ".*",
					Filter: models.FirewallFilter{
						Hostname: "",
						Tags:     []string{"editedtag"},
					},
				},
			},
			fixtures: []string{fixtures.FixtureFirewallRules},
			expected: Expected{
				rule: &models.FirewallRule{
					ID:       "6504b7bd9b6c4a63a9ccc053",
					TenantID: "00000000-0000-4000-0000-000000000000",
					FirewallRuleFields: models.FirewallRuleFields{
						Priority: 1,
						Action:   "deny",
						Active:   true,
						SourceIP: ".*",
						Username: ".*",
						Filter: models.FirewallFilter{
							Hostname: "",
							Tags:     []string{"editedtag"},
						},
					},
				},
				err: nil,
			},
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.id != "" {
				objID, err := primitive.ObjectIDFromHex(tc.id)
				if err != nil {
					t.Fatalf("failed to convert ID to ObjectID: %v", err)
				}

				doc := bson.M{
					"_id":       objID,
					"tenant_id": "00000000-0000-4000-0000-000000000000",
					"priority":  tc.expected.rule.Priority,
					"action":    tc.expected.rule.Action,
					"active":    tc.expected.rule.Active,
					"source_ip": tc.expected.rule.SourceIP,
					"username":  tc.expected.rule.Username,
					"filter": bson.M{
						"hostname": tc.expected.rule.Filter.Hostname,
						"tags":     tc.expected.rule.Filter.Tags,
					},
				}
				if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			rule, err := mongostore.FirewallRuleUpdate(context.TODO(), tc.id, tc.rule)
			assert.Equal(t, tc.expected, Expected{rule: rule, err: err})
		})
	}
}

func TestFirewallRuleDelete(t *testing.T) {
	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when rule is not found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{},
			expected:    FromMongoError(nil),
		},
		{
			description: "succeeds when rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtures.FixtureFirewallRules},
			expected:    nil,
		},
	}

	ctx := context.TODO()

	mongostore := GetMongoStore()
	fixtures.Init(mongoHost, "test")

	collection := mongostore.db.Collection("firewall_rules")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			if tc.id != "" {
				objID, err := primitive.ObjectIDFromHex(tc.id)
				if err != nil {
					t.Fatalf("failed to convert ID to ObjectID: %v", err)
				}

				doc := bson.M{
					"_id": objID,
				}

				if err := dbtest.InsertMockData(ctx, collection, []interface{}{doc}); err != nil {
					t.Fatalf("failed to insert documents: %v", err)
				}
			}

			err := mongostore.FirewallRuleDelete(context.TODO(), tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}
