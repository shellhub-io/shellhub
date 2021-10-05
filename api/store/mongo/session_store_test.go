package mongo

import (
	"testing"

	"github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestSessionCreate(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestSessionGet(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	s, err = mongostore.SessionGet(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.NotEmpty(t, s)
}

func TestSessionList(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	sessions, count, err := mongostore.SessionList(data.Context, paginator.Query{Page: -1, PerPage: -1})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, sessions)
}

func TestSessionSetAuthenticated(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionSetAuthenticated(data.Context, models.UID(data.Device.UID), true)
	assert.NoError(t, err)
}

func TestSessionSetRecorded(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionSetRecorded(data.Context, models.UID(data.Session.UID), false)
	assert.NoError(t, err)

	returnedSession, err := mongostore.SessionGet(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.Equal(t, returnedSession.Recorded, false)

	err = mongostore.SessionSetRecorded(data.Context, models.UID(data.Session.UID), true)
	assert.NoError(t, err)

	returnedSession, err = mongostore.SessionGet(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.Equal(t, returnedSession.Recorded, true)
}

func TestSessionKeepAlive(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionSetLastSeen(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
}

func TestSessionDeleteActives(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionDeleteActives(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
}

func TestSessionCreateRecordFrame(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionCreateRecordFrame(data.Context, models.UID(data.Session.UID), &data.RecordedSession)
	assert.NoError(t, err)
}

func TestSessionGetRecordFrame(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionCreateRecordFrame(data.Context, models.UID(data.Session.UID), &data.RecordedSession)
	assert.NoError(t, err)

	recorded, count, err := mongostore.SessionGetRecordFrame(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
}

func TestSessionDeleteRecordFrame(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s1, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s1)

	session2 := data.Session
	session2.Username = "user2"
	session2.UID = "uid2"

	s2, err := mongostore.SessionCreate(data.Context, session2)
	assert.NoError(t, err)
	assert.NotEmpty(t, s2)

	recordSession2 := data.RecordedSession
	recordSession2.UID = models.UID(session2.UID)
	recordSession2.TenantID = "2"

	err = mongostore.SessionCreateRecordFrame(data.Context, models.UID(data.Session.UID), &data.RecordedSession)
	assert.NoError(t, err)

	err = mongostore.SessionCreateRecordFrame(data.Context, models.UID(session2.UID), &recordSession2)
	assert.NoError(t, err)

	recorded, count, err := mongostore.SessionGetRecordFrame(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)

	err = mongostore.SessionDeleteRecordFrame(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)

	recorded, count, err = mongostore.SessionGetRecordFrame(data.Context, models.UID(data.Session.UID))
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, recorded)

	recorded, count, err = mongostore.SessionGetRecordFrame(data.Context, models.UID(session2.UID))
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.NotEmpty(t, recorded)
}

func TestSessionUpdateDeviceUID(t *testing.T) {
	data := initData()

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())

	_, err := mongostore.NamespaceCreate(data.Context, &data.Namespace)
	assert.NoError(t, err)

	err = mongostore.DeviceCreate(data.Context, data.Device, "hostname")
	assert.NoError(t, err)

	s, err := mongostore.SessionCreate(data.Context, data.Session)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	err = mongostore.SessionUpdateDeviceUID(data.Context, models.UID(data.Device.UID), models.UID("newUID"))
	assert.NoError(t, err)
}
