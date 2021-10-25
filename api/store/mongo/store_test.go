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
	"github.com/shellhub-io/shellhub/pkg/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

type Data struct {
	User              models.User
	Namespace         models.Namespace
	FirewallRule      models.FirewallRule
	Device            models.Device
	Subscription      models.Billing
	PublicKey         models.PublicKey
	Session           models.Session
	RecordedSession   models.RecordedSession
	DeviceAuthRequest models.DeviceAuthRequest
	Context           context.Context
}

func TestStoreGetStats(t *testing.T) {
	db := dbtest.DBServer{}
	defer db.Stop()

	ctx := context.TODO()
	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	user := models.User{UserData: models.UserData{Name: "name", Username: "username", Email: "email"}, UserPassword: models.UserPassword{Password: "password"}}
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

func TestStoreLoadLicense(t *testing.T) {
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

func TestStoreSaveLicense(t *testing.T) {
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

func initData() Data {
	authReq := models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))

	return Data{
		models.User{
			UserData: models.UserData{
				Name:     "user",
				Username: "username",
				Email:    "user@shellhub.io",
			},
			UserPassword: models.UserPassword{
				Password: "password",
			},
			ID: "1",
		},
		models.Namespace{
			Name:     "namespace",
			Owner:    "owner",
			TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			Members: []models.Member{
				{
					ID:   "owner",
					Type: authorizer.MemberTypeOwner,
				},
			},
			MaxDevices: -1,
			Settings:   &models.NamespaceSettings{SessionRecord: true},
		},
		models.FirewallRule{
			FirewallRuleFields: models.FirewallRuleFields{
				Priority: 1,
				Action:   "allow",
				Active:   true,
				SourceIP: ".*",
				Username: ".*",
				Hostname: ".*",
			},
		},
		models.Device{
			UID:      hex.EncodeToString(uid[:]),
			Identity: authReq.Identity,
			TenantID: authReq.TenantID,
			LastSeen: clock.Now(),
		},
		models.Billing{
			SubscriptionID:   "subc_1111x",
			CurrentPeriodEnd: time.Date(2021, time.Month(6), 21, 1, 10, 30, 0, time.UTC),
			PriceID:          "pid_11x",
			Active:           true,
			State:            "pending",
		},
		models.PublicKey{
			Data:            []byte("teste"),
			Fingerprint:     "fingerprint",
			TenantID:        "tenant1",
			PublicKeyFields: models.PublicKeyFields{Name: "teste1", Hostname: ".*"},
		},
		models.Session{
			Username:      "username",
			UID:           "uid",
			TenantID:      "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			DeviceUID:     models.UID(hex.EncodeToString(uid[:])),
			IPAddress:     "0.0.0.0",
			Authenticated: true,
		},
		models.RecordedSession{
			UID:      models.UID("uid"),
			Message:  "message",
			TenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			Time:     clock.Now(),
			Width:    0,
			Height:   0,
		},
		authReq,
		context.TODO(),
	}
}
