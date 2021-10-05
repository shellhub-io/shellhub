package mongo

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFirewallRuleCreate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)
}

func TestFirewallRuleGet(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	rule, err := mongostore.FirewallRuleGet(data.Context, rules[0].ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestFirewallRuleUpdate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	data.FirewallRule.FirewallRuleFields.Action = "deny"

	rule, err := mongostore.FirewallRuleUpdate(
		data.Context,
		rules[0].ID,
		models.FirewallRuleUpdate{
			FirewallRuleFields: data.FirewallRule.FirewallRuleFields,
		},
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, rule)
}

func TestFirewallRuleDelete(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)

	err = mongostore.FirewallRuleDelete(data.Context, rules[0].ID)
	assert.NoError(t, err)
}

func TestFirewallRulesList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, count, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, rules)
}
