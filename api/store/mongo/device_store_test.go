package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceCreate(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
}

func TestDeviceGet(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.DeviceGet(ctx, models.UID(device.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceRename(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
	err = mongostore.DeviceRename(ctx, models.UID(device.UID), "newHostname")
	assert.NoError(t, err)
}

func TestDeviceLookup(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "device")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateStatus(ctx, models.UID(device.UID), "accepted")
	assert.NoError(t, err)

	d, err := mongostore.DeviceLookup(ctx, "name", "device")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceUpdateStatus(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "device")
	assert.NoError(t, err)

	err = mongostore.DeviceUpdateStatus(ctx, models.UID(device.UID), "accepted")
	assert.NoError(t, err)
}

func TestDeviceSetOnline(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
	err = mongostore.DeviceSetOnline(ctx, models.UID(device.UID), true)
	assert.NoError(t, err)
}

func TestDeviceGetByMac(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.DeviceGetByMac(ctx, "mac", "tenant", "pending")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceGetByName(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "hostname")
	assert.NoError(t, err)
	d, err := mongostore.DeviceGetByName(ctx, "hostname", "tenant")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDeviceGetByUID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)
	d, err := mongostore.DeviceGetByUID(ctx, models.UID(device.UID), "tenant")
	assert.NoError(t, err)
	assert.NotEmpty(t, d)
}

func TestDevicesList(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	device := models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: clock.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)

	devices, count, err := mongostore.DeviceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, devices)
}

func TestDeviceListByUsage(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	devices := make([]models.Device, 0)
	sessions := make([]models.Session, 0)

	quantities := []int{10, 5, 3, 1, 1, 0, 0, 0}

	for i, q := range quantities {
		devices = append(devices, models.Device{
			UID:      fmt.Sprintf("%s%d", "uid", i+1),
			TenantID: "tenant",
			Status:   "accepted",
		})
		for j := 0; j < q; j++ {
			sessions = append(sessions, models.Session{
				UID:       fmt.Sprintf("%s%d", "uid", j),
				TenantID:  "tenant",
				DeviceUID: models.UID(fmt.Sprintf("%s%d", "uid", i+1)),
			})
		}
	}

	sessionsInterfaces := make([]interface{}, len(sessions))
	devicesInterfaces := make([]interface{}, len(devices))

	for i, v := range sessions {
		sessionsInterfaces[i] = v
	}

	for i, v := range devices {
		devicesInterfaces[i] = v
	}

	_, _ = db.Client().Database("test").Collection("sessions").InsertMany(ctx, sessionsInterfaces)
	_, _ = db.Client().Database("test").Collection("devices").InsertMany(ctx, devicesInterfaces)

	devices, err = mongostore.DeviceListByUsage(ctx, namespace.TenantID)
	expectedUIDs := []string{"uid1", "uid2", "uid3"}

	assert.NoError(t, err)
	assert.Equal(t, len(expectedUIDs), len(devices))

	for i, device := range devices {
		assert.Equal(t, expectedUIDs[i], device.UID)
	}
}

func TestDeviceChoice(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	devices := make([]models.Device, 0)

	devicesInterfaces := make([]interface{}, 5)

	for i := 0; i < 5; i++ {
		devices = append(devices, models.Device{
			UID:      fmt.Sprintf("%s%d", "uid", i+1),
			TenantID: "tenant",
			Status:   "accepted",
		})
	}

	for i, v := range devices {
		devicesInterfaces[i] = v
	}

	_, err = db.Client().Database("test").Collection("devices").InsertMany(ctx, devicesInterfaces)
	assert.NoError(t, err)

	err = mongostore.DeviceChoice(ctx, namespace.TenantID, []string{"uid1", "uid2", "uid5"})
	assert.NoError(t, err)

	devices, _, err = mongostore.DeviceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, "", "last_seen", "asc")
	assert.NoError(t, err)

	pending := make([]string, 0)

	expected := []string{"uid3", "uid4"}

	for _, dev := range devices {
		if dev.Status == "pending" {
			pending = append(pending, dev.UID)
		}
	}

	assert.Equal(t, expected, pending)
}

func TestDeviceCreateTag(t *testing.T) {
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

	err = mongostore.DeviceCreateTag(ctx, models.UID(device.UID), "device1")
	assert.NoError(t, err)

	err = mongostore.DeviceCreateTag(ctx, models.UID(device.UID), "device2")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(ctx, models.UID(device.UID), "tenant")
	assert.NoError(t, err)
	assert.Equal(t, d.Tags, tags)
}

func TestDeviceDeleteTag(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device := models.Device{
		UID:      "1",
		TenantID: "tenant",
		Tags: []string{
			"device1",
			"device2",
			"device3",
		},
	}

	_, err := db.Client().Database("test").Collection("devices").InsertOne(ctx, &device)
	assert.NoError(t, err)

	err = mongostore.DeviceDeleteTag(ctx, models.UID(device.UID), "device2")
	assert.NoError(t, err)

	d, err := mongostore.DeviceGetByUID(ctx, models.UID(device.UID), "tenant")
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

func TestDeviceListTag(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	device1 := models.Device{
		UID:      "1",
		TenantID: "tenant",
		Tags: []string{
			"device1",
			"device5",
			"device3",
		},
	}

	device2 := models.Device{
		UID:      "1",
		TenantID: "tenant",
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

	_, count, err := mongostore.DeviceListTag(ctx)
	assert.NoError(t, err)
	assert.Equal(t, count, 5)
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

	err = mongostore.DeviceDeleteAllTags(ctx, "tenant1", "device1")
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
