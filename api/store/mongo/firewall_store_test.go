package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFirewallRuleCreate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
}

func TestFirewallRuleGet(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	rules, count, err := mongostore.FirewallRuleList(ctx, paginator.Query{Page: -1, PerPage: -1})

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	rule, err := mongostore.FirewallRuleGet(ctx, rules[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestFirewallRuleUpdate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	rule, err := mongostore.FirewallRuleUpdate(ctx, rules[0].ID, models.FirewallRuleUpdate{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 2,
			Action:   "deny",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestFirewallRuleDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)
	rules, count, err := mongostore.FirewallRuleList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	err = mongostore.FirewallRuleDelete(ctx, rules[0].ID)
	assert.NoError(t, err)
}

func TestFirewallRulesList(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(ctx, &models.FirewallRule{
		FirewallRuleFields: models.FirewallRuleFields{
			Priority: 1,
			Action:   "allow",
			Active:   true,
			SourceIP: ".*",
			Username: ".*",
			Hostname: ".*",
		},
	})
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)
}
