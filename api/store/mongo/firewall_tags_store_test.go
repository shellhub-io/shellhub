package mongo

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestFirewallRuleAddTag(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateTag(data.Context, models.UID(data.Device.UID), []string{"tag1", "tag2"})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, _, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleAddTag(data.Context, rules[0].ID, "tag1")
	assert.NoError(t, err)

	rules, _, err = mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	assert.Equal(t, []string{"tag1"}, rules[0].Filter.Tags)
}

func TestFirewallRuleRemoveTag(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateTag(data.Context, models.UID(data.Device.UID), []string{"tag1", "tag2"})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, _, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleAddTag(data.Context, rules[0].ID, "tag1")
	assert.NoError(t, err)

	rules, _, err = mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	assert.Equal(t, []string{"tag1"}, rules[0].Filter.Tags)

	err = mongostore.FirewallRuleRemoveTag(data.Context, rules[0].ID, "tag1")
	assert.NoError(t, err)

	rules, _, err = mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	assert.Equal(t, []string{}, rules[0].Filter.Tags)
}

func TestFirewallRuleUpdateTag(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateTag(data.Context, models.UID(data.Device.UID), []string{"tag1", "tag2"})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleCreate(data.Context, &data.FirewallRule)
	assert.NoError(t, err)

	rules, _, err := mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	err = mongostore.FirewallRuleUpdateTags(data.Context, rules[0].ID, []string{"tag1", "tag2"})
	assert.NoError(t, err)

	rules, _, err = mongostore.FirewallRuleList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)

	assert.Equal(t, []string{"tag1", "tag2"}, rules[0].Filter.Tags)
}
