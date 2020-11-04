package mongo

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrWrongParamsType = errors.New("wrong parameters type")

type Store struct {
	db *mongo.Database

	store.Store
}

func NewStore(db *mongo.Database) *Store {
	return &Store{db: db}
}

func (s *Store) ListDevices(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error) {
	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, err
	}

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

	if status != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"status": status,
			},
		}}, query...)
	}

	orderVal := map[string]int{
		"asc":  1,
		"desc": -1,
	}

	if sort != "" {
		query = append(query, bson.M{
			"$sort": bson.M{sort: orderVal[order]},
		})
	} else {
		query = append(query, bson.M{
			"$sort": bson.M{"last_seen": -1},
		})
	}

	// Apply filters if any
	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("devices"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	devices := make([]models.Device, 0)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return devices, count, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		device := new(models.Device)
		err = cursor.Decode(&device)
		if err != nil {
			return devices, count, err
		}
		devices = append(devices, *device)
	}

	return devices, count, err
}

func (s *Store) GetDevice(ctx context.Context, uid models.UID) (*models.Device, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	device := new(models.Device)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) DeleteDevice(ctx context.Context, uid models.UID) error {
	if _, err := s.db.Collection("devices").DeleteOne(ctx, bson.M{"uid": uid}); err != nil {
		return err
	}

	if _, err := s.db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": uid}); err != nil {
		return err
	}

	_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})
	return err
}

func (s *Store) AddDevice(ctx context.Context, d models.Device, hostname string) error {
	mac := strings.Replace(d.Identity.MAC, ":", "-", -1)
	if hostname == "" {
		hostname = mac
	}

	q := bson.M{
		"$setOnInsert": bson.M{
			"name":   hostname,
			"status": "pending",
		},
		"$set": d,
	}
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": d.UID}, q, opts)
	return err
}

func (s *Store) RenameDevice(ctx context.Context, uid models.UID, name string) error {
	if _, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"name": name}}); err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error) {
	user := new(models.User)
	if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": namespace}).Decode(&user); err != nil {
		return nil, err
	}

	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": user.TenantID, "name": name, "status": "accepted"}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return err
	}

	if !online {
		_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})
		return err
	}
	device.LastSeen = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"last_seen": device.LastSeen}}, opts)
	if err != nil {
		return err
	}

	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: time.Now(),
		Status:   device.Status,
	}
	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return err
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"status": status}}, opts)
	if err != nil {
		return err
	}
	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: time.Now(),
		Status:   status,
	}
	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"started_at": -1,
			},
		},

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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("sessions"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	sessions := make([]models.Session, 0)
	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		session := new(models.Session)
		err = cursor.Decode(&session)
		if err != nil {
			return sessions, count, err
		}

		device, err := s.GetDevice(ctx, session.DeviceUID)
		if err != nil {
			return sessions, count, err
		}

		session.Device = device
		sessions = append(sessions, *session)
	}

	return sessions, count, err
}

func (s *Store) GetSession(ctx context.Context, uid models.UID) (*models.Session, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	session := new(models.Session)

	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&session)
	if err != nil {
		return nil, err
	}

	device, err := s.GetDevice(ctx, session.DeviceUID)
	if err != nil {
		return nil, err
	}

	session.Device = device

	return session, nil
}

func (s *Store) SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error {
	_, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"authenticated": authenticated}})
	return err
}

func (s *Store) CreateSession(ctx context.Context, session models.Session) (*models.Session, error) {
	session.StartedAt = time.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := s.GetDevice(ctx, session.DeviceUID)
	if err != nil {
		return nil, err
	}

	session.TenantID = device.TenantID

	if _, err := s.db.Collection("sessions").InsertOne(ctx, &session); err != nil {
		return nil, err
	}

	as := &models.ActiveSession{
		UID:      models.UID(session.UID),
		LastSeen: session.StartedAt,
	}

	if _, err := s.db.Collection("active_sessions").InsertOne(ctx, &as); err != nil {
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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	onlineDevices, err := aggregateCount(ctx, s.db.Collection("connected_devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}
	query = append([]bson.M{{
		"$match": bson.M{
			"status": "accepted",
		},
	}}, query...)

	registeredDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "pending",
		},
	}}, query...)

	pendingDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
	}

	query = []bson.M{
		{"$count": "count"},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append([]bson.M{{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		}}, query...)
	}

	query = append([]bson.M{{
		"$match": bson.M{
			"status": "rejected",
		},
	}}, query...)

	rejectedDevices, err := aggregateCount(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, err
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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	activeSessions, err := aggregateCount(ctx, s.db.Collection("sessions"), query)
	if err != nil {
		return nil, err
	}

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}, nil
}

func (s *Store) KeepAliveSession(ctx context.Context, uid models.UID) error {
	session := models.Session{}

	err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session)
	if err != nil {
		return err
	}

	session.LastSeen = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err = s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return err
	}

	activeSession := &models.ActiveSession{
		UID:      uid,
		LastSeen: time.Now(),
	}

	if _, err := s.db.Collection("active_sessions").InsertOne(ctx, &activeSession); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeactivateSession(ctx context.Context, uid models.UID) error {
	session := new(models.Session)
	if err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session); err != nil {
		return err
	}

	session.LastSeen = time.Now()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return err
	}

	_, err = s.db.Collection("active_sessions").DeleteMany(ctx, bson.M{"uid": session.UID})
	return err
}
func (s *Store) RecordSession(ctx context.Context, uid models.UID, recordMessage string, width, height int) error {
	record := new(models.RecordedSession)
	session, _ := s.GetSession(ctx, uid)
	record.UID = uid
	record.Message = recordMessage
	record.Width = width
	record.Height = height
	record.TenantID = session.TenantID
	record.Time = time.Now()

	if _, err := s.db.Collection("recorded_sessions").InsertOne(ctx, &record); err != nil {
		return err
	}

	if _, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"recorded": true}}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Store) GetUserByTenant(ctx context.Context, tenant string) (*models.User, error) {
	user := new(models.User)
	if err := s.db.Collection("users").FindOne(ctx, bson.M{"tenant_id": tenant}).Decode(&user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) GetDeviceByMac(ctx context.Context, mac, tenant, status string) (*models.Device, error) {
	device := new(models.Device)
	if status != "" {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "identity": bson.M{"mac": mac}, "status": status}).Decode(&device); err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "identity": bson.M{"mac": mac}}).Decode(&device); err != nil {
			return nil, err
		}
	}
	return device, nil
}

func (s *Store) GetDeviceByName(ctx context.Context, name, tenant string) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "name": name}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) GetDeviceByUID(ctx context.Context, uid models.UID, tenant string) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenant, "uid": uid}).Decode(&device); err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Store) CreateFirewallRule(ctx context.Context, rule *models.FirewallRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	rule.ID = primitive.NewObjectID().Hex()

	if _, err := s.db.Collection("firewall_rules").InsertOne(ctx, &rule); err != nil {
		return err
	}

	return nil
}

func (s *Store) ListFirewallRules(ctx context.Context, pagination paginator.Query) ([]models.FirewallRule, int, error) {
	query := []bson.M{
		{
			"$sort": bson.M{
				"priority": 1,
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("firewall_rules"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	query = append(query, buildPaginationQuery(pagination)...)

	rules := make([]models.FirewallRule, 0)
	cursor, err := s.db.Collection("firewall_rules").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		rule := new(models.FirewallRule)
		err = cursor.Decode(&rule)
		if err != nil {
			return rules, count, err
		}

		rules = append(rules, *rule)
	}

	return rules, count, err
}

func (s *Store) GetFirewallRule(ctx context.Context, id string) (*models.FirewallRule, error) {
	rule := new(models.FirewallRule)
	if err := s.db.Collection("firewall_rules").FindOne(ctx, bson.M{"_id": id}).Decode(&rule); err != nil {
		return nil, err
	}

	return rule, nil
}

func (s *Store) UpdateFirewallRule(ctx context.Context, id string, rule models.FirewallRuleUpdate) (*models.FirewallRule, error) {
	if err := rule.Validate(); err != nil {
		return nil, err
	}

	if _, err := s.db.Collection("firewall_rules").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": rule}); err != nil {
		return nil, err
	}

	r, err := s.GetFirewallRule(ctx, id)
	return r, err
}

func (s *Store) DeleteFirewallRule(ctx context.Context, id string) error {
	if _, err := s.db.Collection("firewall_rules").DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetRecord(ctx context.Context, uid models.UID) ([]models.RecordedSession, int, error) {
	sessionRecord := make([]models.RecordedSession, 0)

	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
	}

	//Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}
	cursor, err := s.db.Collection("recorded_sessions").Aggregate(ctx, query)
	if err != nil {
		return sessionRecord, 0, err
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		record := new(models.RecordedSession)
		err = cursor.Decode(&record)
		if err != nil {
			return sessionRecord, 0, err
		}

		sessionRecord = append(sessionRecord, *record)
	}

	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	query = append(query, bson.M{
		"$count": "count",
	})

	count, err := aggregateCount(ctx, s.db.Collection("recorded_sessions"), query)
	if err != nil {
		return nil, 0, err
	}
	return sessionRecord, count, nil
}
func (s *Store) UpdateUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	_, err := s.db.Collection("sessions").UpdateMany(ctx, bson.M{"device_uid": oldUID}, bson.M{"$set": bson.M{"device_uid": newUID}})
	return err
}

func (s *Store) UpdateUser(ctx context.Context, username, email, currentPassword, newPassword, tenant string) error {
	user, err := s.GetUserByTenant(ctx, tenant)

	if err != nil {
		return err
	}

	if username != "" && username != user.Username {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": tenant}, bson.M{"$set": bson.M{"username": username}}); err != nil {
			return err
		}
	}

	if email != "" && email != user.Email {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": tenant}, bson.M{"$set": bson.M{"email": email}}); err != nil {
			return err
		}
	}

	if newPassword != "" && newPassword != currentPassword {
		if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": tenant}, bson.M{"$set": bson.M{"password": newPassword}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UpdateDataUserSecurity(ctx context.Context, sessionRecord bool, tenant string) error {
	_, err := s.GetUserByTenant(ctx, tenant)

	if err != nil {
		return err
	}

	if _, err := s.db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": tenant}, bson.M{"$set": bson.M{"session_record": sessionRecord}}); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetDataUserSecurity(ctx context.Context, tenant string) (bool, error) {
	_, err := s.GetUserByTenant(ctx, tenant)

	if err != nil {
		return false, err
	}

	var status struct {
		SessionRecord bool `json:"session_record" bson:"session_record"`
	}

	if err := s.db.Collection("users").FindOne(ctx, bson.M{"tenant_id": tenant}).Decode(&status); err != nil {
		return false, err
	}

	return status.SessionRecord, nil
}

func buildFilterQuery(filters []models.Filter) ([]bson.M, error) {
	var queryMatch []bson.M
	var queryFilter []bson.M

	for _, filter := range filters {
		switch filter.Type {
		case "property":
			var property bson.M
			params, ok := filter.Params.(*models.PropertyParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Operator {
			case "like":
				property = bson.M{"$regex": params.Value, "$options": "i"}
			case "eq":
				property = bson.M{"$eq": params.Value}
			case "bool":
				operator, _ := strconv.ParseBool(params.Value)
				property = bson.M{"$eq": operator}

			case "gt":
				property = bson.M{"$gt": params.Value}
			}

			queryFilter = append(queryFilter, bson.M{
				params.Name: property,
			})
		case "int_property":
			var property bson.M
			params, ok := filter.Params.(*models.IntParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Operator {
			case "eq":
				property = bson.M{"$eq": params.Value}

			case "gt":
				property = bson.M{"$gt": params.Value}
			}

			queryFilter = append(queryFilter, bson.M{
				params.Name: property,
			})

		case "operator":
			var operator string
			params, ok := filter.Params.(*models.OperatorParams)
			if !ok {
				return nil, ErrWrongParamsType
			}

			switch params.Name {
			case "and":
				operator = "$and"
			case "or":
				operator = "$or"
			}

			queryMatch = append(queryMatch, bson.M{
				"$match": bson.M{operator: queryFilter},
			})

			queryFilter = nil
		}
	}

	if len(queryFilter) > 0 {
		queryMatch = append(queryMatch, bson.M{
			"$match": bson.M{"$or": queryFilter},
		})
	}

	return queryMatch, nil
}

func (s *Store) ListUsers(ctx context.Context, pagination paginator.Query, filters []models.Filter, export bool) ([]models.User, int, error) {
	queryMatch, err := buildFilterQuery(filters)
	query := []bson.M{}
	if export {
		query = []bson.M{

			{
				"$lookup": bson.M{
					"from":         "devices",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "devices",
				},
			},
			{
				"$lookup": bson.M{
					"from":         "sessions",
					"localField":   "devices.uid",
					"foreignField": "device_uid",
					"as":           "sessions",
				},
			},
			{
				"$project": bson.M{
					"name":      1,
					"email":     1,
					"username":  1,
					"password":  1,
					"tenant_id": 1,
					"devices": bson.M{
						"$size": "$devices",
					},
					"sessions": bson.M{
						"$size": "$sessions",
					},
				},
			},
		}
	}

	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := append(query, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("users"), queryCount)
	if err != nil {
		return nil, 0, err
	}

	if pagination.Page != 0 && pagination.PerPage != 0 && !export {
		query = append(query, buildPaginationQuery(pagination)...)
	}

	users := make([]models.User, 0)
	cursor, err := s.db.Collection("users").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		user := new(models.User)
		err = cursor.Decode(&user)
		if err != nil {
			return users, count, err
		}

		users = append(users, *user)
	}

	return users, count, err
}

func (s *Store) CreateUser(ctx context.Context, user *models.User) error {
	_, err := s.db.Collection("users").InsertOne(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			return store.ErrDuplicateEmail
		}
	}

	return err
}

func (s *Store) LoadLicense(ctx context.Context) (*models.License, error) {
	findOpts := options.FindOne()
	findOpts.SetSort(bson.M{"created_at": -1})

	license := new(models.License)
	if err := s.db.Collection("licenses").FindOne(ctx, bson.M{}, findOpts).Decode(&license); err != nil {
		return nil, err
	}

	return license, nil
}

func (s *Store) SaveLicense(ctx context.Context, license *models.License) error {
	_, err := s.db.Collection("licenses").InsertOne(ctx, license)
	return err
}

func (s *Store) GetPublicKey(ctx context.Context, fingerprint string) (*models.PublicKey, error) {
	pubKey := new(models.PublicKey)
	if err := s.db.Collection("public_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&pubKey); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return pubKey, nil
}

func (s *Store) CreatePrivateKey(ctx context.Context, key *models.PrivateKey) error {
	_, err := s.db.Collection("private_keys").InsertOne(ctx, key)
	return err
}

func (s *Store) GetPrivateKey(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	privKey := new(models.PrivateKey)
	if err := s.db.Collection("private_keys").FindOne(ctx, bson.M{"fingerprint": fingerprint}).Decode(&privKey); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return privKey, nil
}

func buildPaginationQuery(pagination paginator.Query) []bson.M {
	if pagination.PerPage == -1 {
		return nil
	}

	return []bson.M{
		bson.M{"$skip": pagination.PerPage * (pagination.Page - 1)},
		bson.M{"$limit": pagination.PerPage},
	}
}

func EnsureIndexes(db *mongo.Database) error {
	mod := mongo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(true),
	}
	_, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"last_seen", 1}},
		Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
	}
	_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(false),
	}
	_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(true),
	}
	_, err = db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"last_seen", 1}},
		Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
	}
	_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(false),
	}
	_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"username", 1}},
		Options: options.Index().SetName("username").SetUnique(true),
	}
	_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}

	mod = mongo.IndexModel{
		Keys:    bson.D{{"tenant_id", 1}},
		Options: options.Index().SetName("tenant_id").SetUnique(true),
	}
	_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		return err
	}
	return nil
}
