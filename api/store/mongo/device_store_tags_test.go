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
