package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/store"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Store struct {
	db *mgo.Database

	store.Store
}

func NewStore(db *mgo.Database) *Store {
	return &Store{db: db}
}

func (s *Store) ListDevices(ctx context.Context) ([]models.Device, error) {
	query := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "connected_devices",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "online",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$addFields": bson.M{
				"online":    bson.M{"$anyElementTrue": []interface{}{"$online"}},
				"namespace": "$namespace.username",
			},
		},
		{
			"$unwind": "$namespace",
		},
	}

	// Only match for the respective tenant if requested
	if tenant := store.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	devices := make([]models.Device, 0)
	if err := s.db.C("devices").Pipe(query).All(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}

func (s *Store) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.C("devices").Find(bson.M{"uid": uid}).One(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) DeleteDevice(ctx context.Context, uid models.UID) error {
	if err := s.db.C("devices").Remove(bson.M{"uid": uid}); err != nil {
		return err
	}

	if err := s.db.C("sessions").Remove(bson.M{"device": uid}); err != nil {
		return err
	}

	return nil
}

func (s *Store) AddDevice(ctx context.Context, d models.Device) error {
	hostname := strings.Replace(d.Identity["mac"], ":", "-", -1)

	q := bson.M{
		"$setOnInsert": bson.M{
			"name": hostname,
		},
		"$set": d,
	}

	_, err := s.db.C("devices").Upsert(bson.M{"uid": d.UID}, q)
	return err
}

func (s *Store) RenameDevice(ctx context.Context, uid models.UID, name string) error {
	return s.db.C("devices").Update(bson.M{"uid": uid}, bson.M{"$set": bson.M{"name": name}})
}

func (s *Store) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	user := new(models.User)
	if err := s.db.C("users").Find(bson.M{"$match": bson.M{"username": namespace}}).One(&user); err != nil {
		return nil, err
	}

	device := new(models.Device)
	if err := s.db.C("devices").Find(bson.M{"tenant_id": user.TenantID, "name": name}).One(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	device := new(models.Device)
	if err := s.db.C("devices").Find(bson.M{"uid": uid}).One(&device); err != nil {
		return err
	}

	if !online {
		_, err := s.db.C("connected_devices").RemoveAll(bson.M{"uid": uid})
		return err
	}

	device.LastSeen = time.Now()

	_, err := s.db.C("devices").Upsert(bson.M{"uid": device.UID}, device)
	if err != nil {
		return err
	}

	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: time.Now(),
	}

	return s.db.C("connected_devices").Insert(&cd)
}

func (s *Store) ListSessions(ctx context.Context) ([]models.Session, error) {
	query := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := store.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	sessions := make([]models.Session, 0)
	if err := s.db.C("sessions").Pipe(query).All(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *Store) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	session.StartedAt = time.Now()
	session.LastSeen = session.StartedAt

	device, err := s.GetDevice(ctx, session.Device)
	if err != nil {
		return nil, err
	}

	session.TenantID = device.TenantID

	if err := s.db.C("sessions").Insert(&session); err != nil {
		return nil, err
	}

	as := &models.ActiveSession{
		UID:      models.UID(session.UID),
		LastSeen: session.StartedAt,
	}

	if err := s.db.C("active_sessions").Insert(&as); err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *Store) GetStats(ctx context.Context) (*models.Stats, error) {
	query := []bson.M{
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
		{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
	}

	// Only match for the respective tenant if requested
	if tenant := store.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	resp := []bson.M{}

	if err := s.db.C("connected_devices").Pipe(query).All(&resp); err != nil {
		return nil, err
	}

	onlineDevices := 0
	if len(resp) > 0 {
		onlineDevices = resp[0]["count"].(int)
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := store.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	resp = []bson.M{}

	if err := s.db.C("devices").Pipe(query).All(&resp); err != nil {
		return nil, err
	}

	registeredDevices := 0
	if len(resp) > 0 {
		registeredDevices = resp[0]["count"].(int)
	}

	query = []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
			},
		},
		{
			"$match": bson.M{
				"active": true,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := store.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	resp = []bson.M{}

	if err := s.db.C("sessions").Pipe(query).All(&resp); err != nil {
		return nil, err
	}

	activeSessions := 0
	if len(resp) > 0 {
		activeSessions = resp[0]["count"].(int)
	}

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		ActiveSessions:    activeSessions,
	}, nil
}

func (s *Store) KeepAliveSession(ctx context.Context, uid models.UID) error {
	session := models.Session{}

	err := s.db.C("sessions").Find(bson.M{"uid": s}).One(&session)
	if err != nil {
		return err
	}

	session.LastSeen = time.Now()

	_, err = s.db.C("sessions").Upsert(bson.M{"uid": session.UID}, session)
	if err != nil {
		return err
	}

	activeSession := &models.ActiveSession{
		UID:      uid,
		LastSeen: time.Now(),
	}

	if err := s.db.C("active_sessions").Insert(&activeSession); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeactivateSession(ctx context.Context, uid models.UID) error {
	session := new(models.Session)
	if err := s.db.C("sessions").Find(bson.M{"uid": uid}).One(&session); err != nil {
		return err
	}

	session.LastSeen = time.Now()

	_, err := s.db.C("sessions").Upsert(bson.M{"uid": session.UID}, session)
	if err != nil {
		return err
	}

	_, err = s.db.C("active_sessions").RemoveAll(bson.M{"uid": session.UID})
	return err
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)
	if err := s.db.C("users").Find(bson.M{"username": username}).One(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByTenant(ctx context.Context, tenant string) (*models.User, error) {
	user := new(models.User)
	if err := s.db.C("users").Find(bson.M{"tenant_id": tenant}).One(&user); err != nil {
		return nil, err
	}
	return user, nil
}
