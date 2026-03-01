package migrate

import (
	"context"

	"github.com/google/uuid" //nolint:depguard // migration package generates UUIDs directly
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoSystem struct {
	Setup          bool             `bson:"setup"`
	Authentication *mongoSystemAuth `bson:"authentication"`
}

type mongoSystemAuth struct {
	Local *mongoSystemAuthLocal `bson:"local"`
	SAML  *mongoSystemAuthSAML  `bson:"saml"`
}

type mongoSystemAuthLocal struct {
	Enabled bool `bson:"enabled"`
}

type mongoSystemAuthSAML struct {
	Enabled bool            `bson:"enabled"`
	Idp     *mongoSystemIdp `bson:"idp"`
	Sp      *mongoSystemSp  `bson:"sp"`
}

type mongoSystemIdp struct {
	EntityID     string              `bson:"entity_id"`
	Binding      *mongoSystemBinding `bson:"binding"`
	Certificates []string            `bson:"certificates"`
	Mappings     map[string]string   `bson:"mappings"`
}

type mongoSystemBinding struct {
	Post      string `bson:"post"`
	Redirect  string `bson:"redirect"`
	Preferred string `bson:"preferred"`
}

type mongoSystemSp struct {
	SignAuthRequests bool   `bson:"sign_auth_requests"`
	Certificate      string `bson:"certificate"`
	PrivateKey       string `bson:"private_key"`
}

func convertSystem(doc mongoSystem) *entity.System {
	e := &entity.System{
		ID:    uuid.New().String(),
		Setup: doc.Setup,
	}

	if doc.Authentication != nil {
		if doc.Authentication.Local != nil {
			e.Authentication.Local.Enabled = doc.Authentication.Local.Enabled
		}
		if doc.Authentication.SAML != nil {
			e.Authentication.SAML.Enabled = doc.Authentication.SAML.Enabled
			if doc.Authentication.SAML.Idp != nil {
				e.Authentication.SAML.Idp.EntityID = doc.Authentication.SAML.Idp.EntityID
				e.Authentication.SAML.Idp.Certificates = doc.Authentication.SAML.Idp.Certificates
				e.Authentication.SAML.Idp.Mappings = doc.Authentication.SAML.Idp.Mappings
				if doc.Authentication.SAML.Idp.Binding != nil {
					e.Authentication.SAML.Idp.Binding.Post = doc.Authentication.SAML.Idp.Binding.Post
					e.Authentication.SAML.Idp.Binding.Redirect = doc.Authentication.SAML.Idp.Binding.Redirect
					e.Authentication.SAML.Idp.Binding.Preferred = doc.Authentication.SAML.Idp.Binding.Preferred
				}
			}
			if doc.Authentication.SAML.Sp != nil {
				e.Authentication.SAML.Sp.SignAuthRequests = doc.Authentication.SAML.Sp.SignAuthRequests
				e.Authentication.SAML.Sp.Certificate = doc.Authentication.SAML.Sp.Certificate
				e.Authentication.SAML.Sp.PrivateKey = doc.Authentication.SAML.Sp.PrivateKey
			}
		}
	}

	return e
}

func (m *Migrator) migrateSystems(ctx context.Context) error {
	cursor, err := m.mongo.Collection("system").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var batch []*entity.System
	for cursor.Next(ctx) {
		var doc mongoSystem
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertSystem(doc))
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
	}

	log.WithField("count", len(batch)).Info("Migrated systems")

	return nil
}
