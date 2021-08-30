package mongo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func TestUpdateDeviceStatus(t *testing.T) {
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

func TestCreateSession(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	s, err := mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestGetSession(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	s, err := mongostore.SessionGet(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestListSessions(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	sessions, count, err := mongostore.SessionList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, sessions)
}

func TestSetSessionAuthenticated(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SessionSetAuthenticated(ctx, models.UID(device.UID), true)
	assert.NoError(t, err)
}

func TestSetSessionRecorded(t *testing.T) {
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
		LastSeen: time.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
		Recorded:      true,
	}

	var status bool

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)

	err = mongostore.SessionSetRecorded(ctx, models.UID(session.UID), status)
	assert.NoError(t, err)

	returnedSession, err := mongostore.SessionGet(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, returnedSession.Recorded, status)

	status = true
	err = mongostore.SessionSetRecorded(ctx, models.UID(session.UID), status)
	assert.NoError(t, err)

	returnedSession, err = mongostore.SessionGet(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, returnedSession.Recorded, session.Recorded)
}

func TestKeepAliveSession(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SessionSetLastSeen(ctx, models.UID(session.UID))
	assert.NoError(t, err)
}

func TestDeactivateSession(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SessionDeleteActives(ctx, models.UID(session.UID))
	assert.NoError(t, err)
}

func TestRecordSession(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		TenantID:      "tenant",
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	recordSession := models.RecordedSession{
		UID:      models.UID(session.UID),
		Message:  "message",
		TenantID: session.TenantID,
		Time:     clock.Now(),
		Width:    0,
		Height:   0,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SessionCreateRecordFrame(ctx, models.UID(session.UID), &recordSession)
	assert.NoError(t, err)
}

func TestGetRecord(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		TenantID:      "tenant",
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	recordSession := models.RecordedSession{
		UID:      models.UID(session.UID),
		Message:  "message",
		TenantID: session.TenantID,
		Time:     clock.Now(),
		Width:    0,
		Height:   0,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	err = mongostore.SessionCreateRecordFrame(ctx, models.UID(session.UID), &recordSession)
	assert.NoError(t, err)
	recorded, count, err := mongostore.SessionGetRecordFrame(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
}

func TestGetUserByUsername(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "owner"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
	assert.Equal(t, u.ID, user.ID)
}

func TestGetUserByEmail(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	u, err := mongostore.UserGetByEmail(ctx, "email")
	assert.NoError(t, err)
	assert.NotEmpty(t, u)
}

func TestGetDeviceByMac(t *testing.T) {
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

func TestSessionDeleteRecordFrame(t *testing.T) {
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
		LastSeen: time.Now(),
	}

	err = mongostore.DeviceCreate(ctx, device, "")
	assert.NoError(t, err)

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		TenantID:      "tenant",
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	recordSession := models.RecordedSession{
		UID:      models.UID(session.UID),
		Message:  "message",
		TenantID: session.TenantID,
		Time:     clock.Now(),
		Width:    0,
		Height:   0,
	}

	session2 := models.Session{
		Username:      "user",
		UID:           "uid2",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		TenantID:      "tenant2",
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	recordSession2 := models.RecordedSession{
		UID:      models.UID(session2.UID),
		Message:  "message",
		TenantID: session2.TenantID,
		Time:     clock.Now(),
		Width:    0,
		Height:   0,
	}

	_, err = mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	_, err = mongostore.SessionCreate(ctx, session2)
	assert.NoError(t, err)
	err = mongostore.SessionCreateRecordFrame(ctx, models.UID(session.UID), &recordSession)
	assert.NoError(t, err)
	err = mongostore.SessionCreateRecordFrame(ctx, models.UID(session2.UID), &recordSession2)
	assert.NoError(t, err)
	recorded, count, err := mongostore.SessionGetRecordFrame(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
	err = mongostore.SessionDeleteRecordFrame(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	recorded, count, err = mongostore.SessionGetRecordFrame(ctx, models.UID(session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, recorded)
	recorded, count, err = mongostore.SessionGetRecordFrame(ctx, models.UID(session2.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
}

func TestGetDeviceByName(t *testing.T) {
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

func TestGetDeviceByUID(t *testing.T) {
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

func TestCreateFirewallRule(t *testing.T) {
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

func TestGetFirewallRule(t *testing.T) {
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

func TestUpdateFirewallRule(t *testing.T) {
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

func TestDeleteFirewallRule(t *testing.T) {
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

func TestListDevices(t *testing.T) {
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

func TestListFirewallRules(t *testing.T) {
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

func TestUpdateUID(t *testing.T) {
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
	err = mongostore.SessionUpdateDeviceUID(ctx, models.UID(device.UID), models.UID("newUID"))
	assert.NoError(t, err)
}

func TestUserUpdateData(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	userNewData := models.User{
		ID:       objID,
		Name:     "New Name",
		Username: "newusername",
		Password: "password",
		Email:    "new@email.com",
	}

	err = mongostore.UserUpdateData(ctx, &userNewData, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
	assert.Equal(t, us, &userNewData)
	assert.NoError(t, err)
}

func TestUserGetByID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	namespacesOwner := []models.Namespace{
		{
			Name:     "namespace1",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant1",
		},
		{
			Name:     "namespace2",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant2",
		},
		{
			Name:     "namespace3",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant3",
		},
		{
			Name:     "namespace4",
			Owner:    "60af83d418d2dc3007cd445c",
			TenantID: "tenant4",
		},
	}

	namespacesNotOwner := []models.Namespace{
		{
			Name:     "namespace18",
			Owner:    "60af83d418d2dc3007cd445d",
			TenantID: "tenant18",
		},
		{
			Name:     "namespace19",
			Owner:    "60af83d418d2dc3007cd445e",
			TenantID: "tenant19",
		},
		{
			Name:     "namespace20",
			Owner:    "60af83d418d2dc3007cd445f",
			TenantID: "tenant20",
		},
		{
			Name:     "namespace21",
			Owner:    "60af83d418d2dc3007cd4451",
			TenantID: "tenant21",
		},
		{
			Name:     "namespace22",
			Owner:    "60af83d418d2dc3007cd4452",
			TenantID: "tenant22",
		},
		{
			Name:     "namespace23",
			Owner:    "60af83d418d2dc3007cd4453",
			TenantID: "tenant23",
		},
		{
			Name:     "namespace24",
			Owner:    "60af83d418d2dc3007cd4454",
			TenantID: "tenant24",
		},
		{
			Name:     "namespace25",
			Owner:    "60af83d418d2dc3007cd4455",
			TenantID: "tenant25",
		},
		{
			Name:     "namespace26",
			Owner:    "060af83d418d2dc3007cd4456",
			TenantID: "tenant26",
		},
	}

	namespaces := append(namespacesOwner, namespacesNotOwner...)
	nss := make([]interface{}, len(namespaces))

	for i, v := range namespaces {
		nss[i] = v
	}

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertMany(ctx, nss)
	assert.NoError(t, err)

	us, countNs, err := mongostore.UserGetByID(ctx, user.ID, true)
	assert.NoError(t, err)
	assert.Equal(t, countNs, len(namespacesOwner))
	assert.Equal(t, us, &user)

	us, countNs, err = mongostore.UserGetByID(ctx, user.ID, false)
	assert.NoError(t, err)
	assert.Equal(t, countNs, 0)
	assert.Equal(t, us, &user)
}

func TestUserUpdatePassword(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	newPassword := "password2"

	err = mongostore.UserUpdatePassword(ctx, newPassword, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
	assert.Equal(t, us.Password, newPassword)
	assert.NoError(t, err)
}

func TestUpdateUserFromAdmin(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()
	err = mongostore.UserUpdateFromAdmin(ctx, "newName", "newUsername", "newEmail", "password", objID)
	assert.NoError(t, err)
}

func TestUserCreateToken(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", Authenticated: false}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	UserTokenRecover := models.UserTokenRecover{Token: "token", User: objID}

	err = mongostore.UserCreateToken(ctx, &UserTokenRecover)
	assert.NoError(t, err)

	userToken, err := mongostore.UserGetToken(ctx, objID)
	assert.Equal(t, userToken.Token, UserTokenRecover.Token)
	assert.NoError(t, err)
}

func TestUpdateAccountStatus(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", Authenticated: false}

	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	objID := result.InsertedID.(primitive.ObjectID).Hex()

	err = mongostore.UserUpdateAccountStatus(ctx, objID)
	assert.NoError(t, err)

	us, _, err := mongostore.UserGetByID(ctx, objID, false)
	assert.Equal(t, us.Authenticated, true)
	assert.NoError(t, err)
}

func TestGetDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	returnedStatus, err := mongostore.NamespaceGetSessionRecord(ctx, namespace.TenantID)
	assert.Equal(t, returnedStatus, namespace.Settings.SessionRecord)
	assert.NoError(t, err)
}

func TestUpdateDataUserSecurity(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email", ID: "hash1"}
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "a736a52b-5777-4f92-b0b8-e359bf484713", Settings: &models.NamespaceSettings{SessionRecord: true}}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.NamespaceSetSessionRecord(ctx, false, namespace.TenantID)
	assert.NoError(t, err)
}

func TestListUsers(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	userID := result.InsertedID.(primitive.ObjectID).Hex()
	namespace := models.Namespace{Name: "name", Owner: userID, TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, users)
}

func TestListUsersWithFilter(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	result, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace := models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-1", Password: "password", Email: "email-1"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-2", Password: "password", Email: "email-2"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	user = models.User{Name: "name", Username: "username-3", Password: "password", Email: "email-3"}
	result, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	namespace = models.Namespace{Name: "name", Owner: result.InsertedID.(primitive.ObjectID).Hex(), TenantID: "tenant"}
	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	filters := []models.Filter{
		{
			Type:   "property",
			Params: &models.PropertyParams{Name: "namespaces", Operator: "gt", Value: "1"},
		},
	}

	users, count, err := mongostore.UserList(ctx, paginator.Query{Page: -1, PerPage: -1}, filters)
	assert.NoError(t, err)
	assert.Equal(t, len(users), count)
	assert.Equal(t, 2, count)
	assert.NotEmpty(t, users)
}

func TestGetStats(t *testing.T) {
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

	session := models.Session{
		Username:      "user",
		UID:           "uid",
		DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
		IPAddress:     "0.0.0.0",
		Authenticated: true,
	}

	s, err := mongostore.SessionCreate(ctx, session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	stats, err := mongostore.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, stats)
	assert.Equal(t, 0, stats.RegisteredDevices)
	assert.Equal(t, 0, stats.OnlineDevices)
	assert.Equal(t, 1, stats.PendingDevices)
	assert.Equal(t, 0, stats.RejectedDevices)
	assert.Equal(t, 1, stats.ActiveSessions)
}

func TestCreateUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
}

func TestCreateNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)
}

func TestDeleteNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	err = mongostore.NamespaceDelete(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}

func TestGetNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, err = mongostore.NamespaceGet(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
	assert.NoError(t, err)
}

func TestListNamespaces(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	_, count, err := mongostore.NamespaceList(ctx, paginator.Query{Page: -1, PerPage: -1}, nil, false)
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestAddNamespaceUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user_id",
	})
	assert.NoError(t, err)
	err = mongostore.UserCreate(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
		ID:       "user2_id",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestUpdateNamespace(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Name:     "name",
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)

	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		Settings:   &models.NamespaceSettings{SessionRecord: true},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	err = mongostore.NamespaceUpdate(ctx, "tenant", &models.Namespace{
		Name:       "name",
		Settings:   &models.NamespaceSettings{SessionRecord: false},
		MaxDevices: 3,
	})
	assert.NoError(t, err)
}

func TestRemoveNamespaceUser(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.UserCreate(ctx, &models.User{
		Username: "user",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	err = mongostore.UserCreate(ctx, &models.User{
		Username: "user2",
		Email:    "user@shellhub.io",
		Password: "password",
	})
	assert.NoError(t, err)
	_, err = mongostore.NamespaceCreate(ctx, &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	})
	assert.NoError(t, err)

	u, err := mongostore.UserGetByUsername(ctx, "user")
	assert.NoError(t, err)

	_, err = mongostore.NamespaceAddMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)

	_, err = mongostore.NamespaceRemoveMember(ctx, "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", u.ID)
	assert.NoError(t, err)
}

func TestLoadLicense(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.LicenseSave(ctx, &models.License{
		RawData:   []byte("bar"),
		CreatedAt: clock.Now().Local().Truncate(time.Millisecond),
	})
	assert.NoError(t, err)

	license := &models.License{
		RawData:   []byte("foo"),
		CreatedAt: clock.Now().Local().Truncate(time.Millisecond),
	}

	err = mongostore.LicenseSave(ctx, license)
	assert.NoError(t, err)

	loadedLicense, err := mongostore.LicenseLoad(ctx)
	assert.NoError(t, err)

	assert.True(t, license.CreatedAt.Equal(loadedLicense.CreatedAt))

	// decoded value is not in local this won't match with assert.Equal
	loadedLicense.CreatedAt = loadedLicense.CreatedAt.Local()
	assert.Equal(t, license, loadedLicense)
}

func TestSaveLicense(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	err := mongostore.LicenseSave(ctx, &models.License{
		RawData:   []byte("foo"),
		CreatedAt: clock.Now().Truncate(time.Millisecond),
	})
	assert.NoError(t, err)
}

func TestCreatePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}
	err := mongostore.PublicKeyCreate(ctx, newKey)
	assert.NoError(t, err)
}

func TestListPublicKeys(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	_, count, err := mongostore.PublicKeyList(ctx, paginator.Query{Page: -1, PerPage: -1})
	assert.Equal(t, 1, count)
	assert.NoError(t, err)
}

func TestListGetPublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.PublicKeyGet(ctx, key.Fingerprint, key.TenantID)
	assert.NoError(t, err)
	assert.NotEmpty(t, k)
}

func TestUpdatePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	// createdAt := time.Now()
	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}
	updatedKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste2", Hostname: ".*"},
	}
	unexistingKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint2", TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste", Hostname: ".*"},
	}

	update := &models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{Name: "teste2", Hostname: ".*"},
	}

	_, err := db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("public_keys").InsertOne(ctx, key)
	assert.NoError(t, err)

	k, err := mongostore.PublicKeyUpdate(ctx, key.Fingerprint, key.TenantID, update)
	assert.NoError(t, err)
	assert.Equal(t, k, updatedKey)
	_, err = mongostore.PublicKeyUpdate(ctx, unexistingKey.Fingerprint, unexistingKey.TenantID, update)
	assert.EqualError(t, err, store.ErrNoDocuments.Error())
}

func TestDeletePublicKey(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{Name: "name", Username: "username", Password: "password", Email: "email"}
	namespace := models.Namespace{Name: "name", Owner: "owner", TenantID: "tenant"}
	newKey := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", TenantID: "tenant", PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
	}

	_, err := db.Client().Database("test").Collection("public_keys").InsertOne(ctx, newKey)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("users").InsertOne(ctx, user)
	assert.NoError(t, err)

	_, err = db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)
	assert.NoError(t, err)

	err = mongostore.PublicKeyDelete(ctx, newKey.Fingerprint, newKey.TenantID)
	assert.NoError(t, err)
}

func TestBillingUpdateCustomer(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	namespace := models.Namespace{
		TenantID: "tenant",
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, namespace)

	assert.NoError(t, err)

	custID := "cust19x"
	err = mongostore.BillingUpdateCustomer(ctx, &namespace, custID)
	assert.NoError(t, err)

	ns, err := mongostore.NamespaceGet(ctx, namespace.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, custID, ns.Billing.CustomerID)
}

func TestBillingUpdatePaymentID(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	ns := models.Namespace{
		TenantID: "tenant",
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, ns)

	assert.NoError(t, err)

	payID := "pm_89x"
	err = mongostore.BillingUpdatePaymentID(ctx, &ns, payID)
	assert.NoError(t, err)

	namespace, err := mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, payID, namespace.Billing.PaymentMethodID)
}

func TestBillingUpdateSubscription(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
	}
	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	subscription := &models.Billing{
		SubscriptionID:   "subc_1111x",
		CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		PriceID:          "pid_11x",
		Active:           true,
	}

	err = mongostore.BillingUpdateSubscription(ctx, ns, subscription)
	assert.NoError(t, err)

	ns, err = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.NoError(t, err)
	assert.Equal(t, subscription, ns.Billing)
}

func TestBillingUpdatePaymentFailed(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	pf := &models.PaymentFailed{
		Status:  true,
		Details: "invalid",
		Date:    time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		Amount:  47.54,
	}

	ns := &models.Namespace{
		TenantID: "tenant",
	}

	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	_, err = mongostore.BillingUpdatePaymentFailed(ctx, "subs_id", true, pf)
	assert.Error(t, err)

	subsID := "subs_id_1"

	ns2 := &models.Namespace{
		TenantID: "tenant2",
		Billing: &models.Billing{
			SubscriptionID: subsID,
		},
	}

	_, err = mongostore.NamespaceCreate(ctx, ns2)
	assert.NoError(t, err)

	ns2, err = mongostore.BillingUpdatePaymentFailed(ctx, subsID, true, pf)
	assert.NoError(t, err)

	assert.Equal(t, pf, ns2.Billing.PaymentFailed)

	ns2, err = mongostore.BillingUpdatePaymentFailed(ctx, subsID, false, nil)
	assert.NoError(t, err)

	assert.Nil(t, ns2.Billing.PaymentFailed)
}

func TestBillingUpdateSubscriptionPeriodEnd(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	ns := &models.Namespace{
		Name:     "namespace",
		TenantID: "tenant",
		Billing: &models.Billing{
			SubscriptionID:   "subs_id",
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
		},
	}

	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	newDate := time.Date(2021, time.Month(7), 21, 1, 10, 30, 0, time.UTC)

	err = mongostore.BillingUpdateSubscriptionPeriodEnd(ctx, ns.Billing.SubscriptionID, newDate)
	assert.Nil(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.Equal(t, newDate, ns.Billing.CurrentPeriodEnd)
}

func TestBillingUpdateDeviceLimit(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: 3,
		Billing: &models.Billing{
			SubscriptionID:   "subc_1111x",
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
			PriceID:          "pid_11x",
		},
	}
	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	newDeviceLimit := 16

	_, err = mongostore.BillingUpdateDeviceLimit(ctx, "subc_w1x", newDeviceLimit)
	assert.Error(t, err)

	_, err = mongostore.BillingUpdateDeviceLimit(ctx, ns.Billing.SubscriptionID, newDeviceLimit)
	assert.NoError(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.Equal(t, ns.MaxDevices, newDeviceLimit)
}

func TestBillingDeleteCustomer(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	subsID := "sub_1x"
	billing := models.Billing{
		CustomerID:      "cust_111x",
		PaymentMethodID: "pid_111x",
		SubscriptionID:  subsID,
		Active:          true,
	}

	namespace := models.Namespace{
		TenantID: "tenant",
	}

	namespaceBill := models.Namespace{
		TenantID: namespace.TenantID,
		Billing:  &billing,
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, &namespaceBill)

	assert.NoError(t, err)

	ns, _ := mongostore.NamespaceGet(ctx, namespace.TenantID)
	err = mongostore.BillingDeleteCustomer(ctx, ns)
	assert.NoError(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, namespace.TenantID)
	assert.Equal(t, subsID, ns.Billing.SubscriptionID)
}

func TestBillingDeleteSubscription(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	subsID := "subc_1111x"
	ns := &models.Namespace{
		Name:       "namespace",
		Owner:      "owner",
		TenantID:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		Members:    []interface{}{"owner"},
		MaxDevices: -1,
		Billing: &models.Billing{
			SubscriptionID:   subsID,
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
			Active:           true,
			PriceID:          "pid_11x",
		},
	}

	_, err := mongostore.NamespaceCreate(ctx, ns)
	assert.NoError(t, err)

	err = mongostore.BillingDeleteSubscription(ctx, ns.TenantID)
	assert.NoError(t, err)

	ns, _ = mongostore.NamespaceGet(ctx, ns.TenantID)
	assert.Equal(t, false, ns.Billing.Active)
}

func TestBillingRemoveInstance(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	subsID := "sub_1x"
	billing := models.Billing{
		CustomerID:      "cust_111x",
		PaymentMethodID: "pid_111x",
		SubscriptionID:  subsID,
	}

	namespace := models.Namespace{
		TenantID: "tenant",
	}

	namespaceBill := models.Namespace{
		TenantID: namespace.TenantID,
		Billing:  &billing,
	}

	_, err := db.Client().Database("test").Collection("namespaces").InsertOne(ctx, &namespaceBill)

	assert.NoError(t, err)

	_, _ = mongostore.NamespaceGet(ctx, namespace.TenantID)
	err = mongostore.BillingRemoveInstance(ctx, subsID)
	assert.NoError(t, err)

	ns, _ := mongostore.NamespaceGet(ctx, namespace.TenantID)
	assert.Empty(t, ns.Billing)
	assert.Nil(t, ns.Billing)
}

func TestUserDelete(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, _ = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	err = mongostore.UserDelete(ctx, user.ID)
	assert.NoError(t, err)
	_, err = mongostore.UserGetByUsername(ctx, "username")
	assert.Error(t, err, mongo.ErrNoDocuments)
}

func TestUserDetachInfo(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	user := models.User{ID: "60af83d418d2dc3007cd445c", Name: "name", Username: "username", Password: "password", Email: "user@email.com"}

	objID, err := primitive.ObjectIDFromHex(user.ID)

	assert.NoError(t, err)

	_, _ = db.Client().Database("test").Collection("users").InsertOne(ctx, bson.M{
		"_id":      objID,
		"name":     user.Name,
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	})

	namespacesOwner := []*models.Namespace{
		{
			Owner:   user.ID,
			Name:    "ns2",
			Members: []interface{}{user.ID},
		},
		{
			Owner:   user.ID,
			Name:    "ns4",
			Members: []interface{}{user.ID},
		},
	}

	namespacesMember := []*models.Namespace{
		{
			Owner:   "id2",
			Name:    "ns1",
			Members: []interface{}{"id2", user.ID},
		},
		{
			Owner:   "id2",
			Name:    "ns3",
			Members: []interface{}{"id2", user.ID},
		},
		{
			Owner:   "id2",
			Name:    "ns5",
			Members: []interface{}{"id2", user.ID},
		},
	}

	namespaces := append(namespacesOwner, namespacesMember...)
	nss := make([]interface{}, len(namespaces))

	for i, v := range namespaces {
		nss[i] = v
	}

	_, _ = db.Client().Database("test").Collection("namespaces").InsertMany(ctx, nss)

	u, err := mongostore.UserGetByUsername(ctx, "username")
	assert.NoError(t, err)
	assert.Equal(t, user.Username, u.Username)

	namespacesMap, err := mongostore.UserDetachInfo(ctx, user.ID)

	assert.NoError(t, err)
	assert.Equal(t, namespacesMap["member"], namespacesMember)
	assert.Equal(t, namespacesMap["owner"], namespacesOwner)
}
