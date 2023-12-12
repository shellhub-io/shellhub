package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFirewallRuleList(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	type Expected struct {
		rules []models.FirewallRule
		len   int
		err   error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when no firewall rules are found",
			fixtures:    []string{},
			expected: Expected{
				rules: []models.FirewallRule{},
				len:   0,
				err:   nil,
			},
		},
		{
			description: "succeeds when a firewall rule is found",
			fixtures:    []string{fixtures.FirewallRule},
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
								Tags:     []string{"tag1"},
							},
						},
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			rules, count, err := mongostore.FirewallRuleList(ctx, paginator.Query{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{rules: rules, len: count, err: err})
		})
	}
}

func TestFirewallRuleGet(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			id:          "6504b7bd9b6c4a63a9ccc021",
			fixtures:    []string{fixtures.FirewallRule},
			expected: Expected{
				rule: nil,
				err:  store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when firewall rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtures.FirewallRule},
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
							Tags:     []string{"tag1"},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			rule, err := mongostore.FirewallRuleGet(ctx, tc.id)
			assert.Equal(t, tc.expected, Expected{rule: rule, err: err})
		})
	}
}

func TestFirewallRuleUpdate(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

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
			id:          "6504b7bd9b6c4a63a9ccc000",
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
			fixtures: []string{fixtures.FirewallRule},
			expected: Expected{
				rule: nil,
				err:  store.ErrNoDocuments,
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
			fixtures: []string{fixtures.FirewallRule},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			rule, err := mongostore.FirewallRuleUpdate(ctx, tc.id, tc.rule)
			assert.Equal(t, tc.expected, Expected{rule: rule, err: err})
		})
	}
}

func TestFirewallRuleDelete(t *testing.T) {
	ctx := context.TODO()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when rule is not found",
			id:          "6504ac006bf3dbca079f76b1",
			fixtures:    []string{},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtures.FirewallRule},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.FirewallRuleDelete(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}
