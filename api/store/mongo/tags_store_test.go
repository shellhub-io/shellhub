package mongo

import (
	"context"
	"sort"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGetTags(t *testing.T) {
	dbStand := dbtest.DBServer{}
	defer dbStand.Stop()

	mongostoreStand := NewStore(dbStand.Client().Database("test"), cache.NewNullCache())

	db := dbtest.DBServer{Replicaset: true}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device1 := models.Device{
		UID:       "1",
		Namespace: "namespace1",
		TenantID:  "tenant1",
		Tags: []string{
			"device1",
			"device2",
			"device3",
		},
	}

	device2 := models.Device{
		UID:       "2",
		Namespace: "namespace2",
		TenantID:  "tenant2",
		Tags: []string{
			"device4",
			"device5",
			"device6",
		},
	}

	key1 := models.PublicKey{
		Fingerprint: "fingerprint1",
		TenantID:    "tenant1",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"device1", "device2", "device4"},
			},
		},
	}

	rule1 := models.FirewallRule{
		ID:       "rule1",
		TenantID: "tenant1",
		FirewallRuleFields: models.FirewallRuleFields{
			Filter: models.FirewallFilter{
				Tags: []string{"device2", "device5"},
			},
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, &key1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, &rule1)
	assert.NoError(t, err)

	_, _, err = mongostoreStand.TagsGet(ctx, "tenant1")
	assert.Error(t, err)

	tags, count, err := mongostore.TagsGet(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, 5, count)

	sort.Strings(tags) // Guarantee the order for comparison.
	assert.Equal(t, []string{"device1", "device2", "device3", "device4", "device5"}, tags)
}

func TestRenameTag(t *testing.T) {
	dbStand := dbtest.DBServer{}
	defer dbStand.Stop()

	mongostoreStand := NewStore(dbStand.Client().Database("test"), cache.NewNullCache())

	db := dbtest.DBServer{Replicaset: true}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device1 := models.Device{
		UID:      "1",
		TenantID: "tenant1",
		Tags: []string{
			"device1",
			"device2",
			"device3",
		},
	}

	device2 := models.Device{
		UID:      "2",
		TenantID: "tenant2",
		Tags: []string{
			"device1",
			"device2",
			"device3",
		},
	}

	device3 := models.Device{
		UID:      "3",
		TenantID: "tenant1",
		Tags: []string{
			"device1",
			"device2",
			"device3",
		},
	}

	key1 := models.PublicKey{
		Fingerprint: "fingerprint1",
		TenantID:    "tenant1",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"device1", "device4", "device7"},
			},
		},
	}

	rule1 := models.FirewallRule{
		ID:       "rule1",
		TenantID: "tenant1",
		FirewallRuleFields: models.FirewallRuleFields{
			Filter: models.FirewallFilter{
				Tags: []string{"device6", "device7", "device8"},
			},
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device3)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, &key1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, &rule1)
	assert.NoError(t, err)

	err = mongostoreStand.TagRename(ctx, "tenant1", "device2", "device9")
	assert.Error(t, err)

	err = mongostore.TagRename(ctx, "tenant1", "device2", "device9")
	assert.NoError(t, err)

	tags1, _, err := mongostore.TagsGet(ctx, "tenant1")
	sort.Strings(tags1) // Guarantee the order for comparison.
	assert.NoError(t, err)
	assert.Equal(t, []string{"device1", "device3", "device4", "device6", "device7", "device8", "device9"}, tags1)

	tags2, _, err := mongostore.TagsGet(ctx, "tenant2")
	sort.Strings(tags2) // Guarantee the order for comparison.
	assert.NoError(t, err)
	assert.Equal(t, []string{"device1", "device2", "device3"}, tags2)
}

func TestDeleteTag(t *testing.T) {
	dbStand := dbtest.DBServer{}
	defer dbStand.Stop()

	mongostoreStand := NewStore(dbStand.Client().Database("test"), cache.NewNullCache())

	db := dbtest.DBServer{Replicaset: true}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device1 := models.Device{
		UID:       "1",
		Namespace: "namespace1",
		TenantID:  "tenant1",
		Tags: []string{
			"device1",
			"device5",
			"device3",
		},
	}

	device2 := models.Device{
		UID:       "2",
		Namespace: "namespace1",
		TenantID:  "tenant1",
		Tags: []string{
			"device1",
			"device5",
			"device6",
		},
	}

	device3 := models.Device{
		UID:       "3",
		Namespace: "namespace2",
		TenantID:  "tenant2",
		Tags: []string{
			"device1",
			"device5",
			"device6",
		},
	}

	key1 := models.PublicKey{
		Fingerprint: "fingerprint1",
		TenantID:    "tenant1",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"device1", "device4", "device7"},
			},
		},
	}

	rule1 := models.FirewallRule{
		ID:       "rule1",
		TenantID: "tenant1",
		FirewallRuleFields: models.FirewallRuleFields{
			Filter: models.FirewallFilter{
				Tags: []string{"device6", "device7", "device8"},
			},
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device3)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, &key1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("firewall_rules").InsertOne(ctx, &rule1)
	assert.NoError(t, err)

	err = mongostoreStand.TagDelete(ctx, "tenant1", "device1")
	assert.Error(t, err)

	err = mongostore.TagDelete(ctx, "tenant1", "device1")
	assert.NoError(t, err)

	tags1, _, err := mongostore.TagsGet(ctx, "tenant1")
	sort.Strings(tags1) // Guarantee the order for comparison.
	assert.NoError(t, err)
	assert.Equal(t, []string{"device3", "device4", "device5", "device6", "device7", "device8"}, tags1)

	tags2, _, err := mongostore.TagsGet(ctx, "tenant2")
	sort.Strings(tags2) // Guarantee the order for comparison.
	assert.NoError(t, err)
	assert.Equal(t, []string{"device1", "device5", "device6"}, tags2)
}
