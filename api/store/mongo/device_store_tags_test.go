package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreateTag(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(data.Context, models.UID(data.Device.UID), "device1")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(data.Context, models.UID(data.Device.UID), "device2")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(data.Context, models.UID(data.Device.UID), "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
	assert.Equal(t, d.Tags, []string{"device1", "device2"})
}

func TestDeviceDeleteTag(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(data.Context, models.UID(data.Device.UID), "device1")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(data.Context, models.UID(data.Device.UID), "device2")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(data.Context, models.UID(data.Device.UID), "device3")
	assert.NoError(t, err)

	err = mongostore.DeviceRemoveTag(data.Context, models.UID(data.Device.UID), "device2")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(data.Context, models.UID(data.Device.UID), "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
	assert.Equal(t, len(d.Tags), 2)
	assert.Equal(t, d.Tags, []string{"device1", "device3"})
}

func TestDeviceRenameTag(t *testing.T) {
	db := dbtest.DBServer{}
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

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device3)
	assert.NoError(t, err)

	err = mongostore.DeviceRenameTag(ctx, "tenant1", "device2", "device4")
	assert.NoError(t, err)

	d1, err := mongostore.DeviceGetByUID(ctx, models.UID(device1.UID), "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, len(d1.Tags), 3)
	assert.Equal(t, d1.Tags[1], "device4")

	d2, err := mongostore.DeviceGetByUID(ctx, models.UID(device2.UID), "tenant2")
	assert.NoError(t, err)
	assert.Equal(t, len(d2.Tags), 3)
	assert.Equal(t, d2.Tags[1], "device2")

	d3, err := mongostore.DeviceGetByUID(ctx, models.UID(device3.UID), "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, len(d3.Tags), 3)
	assert.Equal(t, d3.Tags[1], "device4")
}

func TestDeviceUpdateTag(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device := models.Device{
		UID:      "1",
		TenantID: "tenant",
		Tags:     []string{},
	}

	tags := []string{
		"device1",
		"device2",
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device)
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateTag(ctx, models.UID(device.UID), tags)
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(ctx, models.UID(device.UID), "tenant")
	assert.NoError(t, err)
	assert.Equal(t, d.Tags, tags)
}

func TestDeviceGetTags(t *testing.T) {
	db := dbtest.DBServer{}
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
		Namespace: "namespace2",
		TenantID:  "tenant2",
		Tags: []string{
			"device4",
			"device5",
			"device6",
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, count, err := mongostore.DeviceGetTags(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, count, 3)
}

func TestDeviceDeleteAllTags(t *testing.T) {
	db := dbtest.DBServer{}
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

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device1)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device2)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("devices").InsertOne(ctx, &device3)
	assert.NoError(t, err)

	err = mongostore.DeviceDeleteTags(ctx, "tenant1", "device1")
	assert.NoError(t, err)

	d1, err := mongostore.DeviceGetByUID(ctx, models.UID(device1.UID), "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, len(d1.Tags), 2)
	assert.Equal(t, d1.Tags, []string{"device5", "device3"})

	d2, err := mongostore.DeviceGetByUID(ctx, models.UID(device2.UID), "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, len(d2.Tags), 2)
	assert.Equal(t, d2.Tags, []string{"device5", "device6"})

	d3, err := mongostore.DeviceGetByUID(ctx, models.UID(device3.UID), "tenant2")
	assert.NoError(t, err)
	assert.Equal(t, len(d3.Tags), 3)
	assert.Equal(t, d3.Tags, device3.Tags)
}
