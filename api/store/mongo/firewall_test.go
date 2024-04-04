package mongo_test

import (
	"context"
	"sort"
	"testing"

	shstore "github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
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
			fixtures:    []string{fixtureFirewallRules},
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
								Tags:     []string{},
							},
						},
					},
					{
						ID:       "3fd759a1ecb64ec5a07c8c0f",
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
			description: "succeeds when firewall rule list is not empty and paginator is different than -1",
			paginator:   query.Paginator{Page: 2, PerPage: 2},
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				rules: []models.FirewallRule{
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
								Tags:     []string{},
							},
						},
					},
					{
						ID:       "3fd759a1ecb64ec5a07c8c0f",
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
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(fr []models.FirewallRule) {
		sort.Slice(fr, func(i, j int) bool {
			return fr[i].ID < fr[j].ID
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			rules, count, err := store.FirewallRuleList(ctx, tc.paginator)

			sort(tc.expected.rules)
			sort(rules)

			assert.Equal(t, tc.expected, Expected{rules: rules, len: count, err: err})
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
			id:          "6504b7bd9b6c4a63a9ccc021",
			fixtures:    []string{fixtureFirewallRules},
			expected: Expected{
				rule: nil,
				err:  shstore.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when firewall rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtureFirewallRules},
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			rule, err := store.FirewallRuleGet(ctx, tc.id)
			assert.Equal(t, tc.expected, Expected{rule: rule, err: err})
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
			fixtures: []string{fixtureFirewallRules},
			expected: Expected{
				rule: nil,
				err:  shstore.ErrNoDocuments,
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
			fixtures: []string{fixtureFirewallRules},
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
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			rule, err := store.FirewallRuleUpdate(ctx, tc.id, tc.rule)
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
			id:          "6504ac006bf3dbca079f76b1",
			fixtures:    []string{},
			expected:    shstore.ErrNoDocuments,
		},
		{
			description: "succeeds when rule is found",
			id:          "6504b7bd9b6c4a63a9ccc053",
			fixtures:    []string{fixtureFirewallRules},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := store.FirewallRuleDelete(ctx, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}
}
