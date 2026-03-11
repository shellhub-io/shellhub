package migrate

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateSystems(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("system").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	if !cursor.Next(ctx) {
		return cursor.Err()
	}

	var doc mongoSystem
	if err := cursor.Decode(&doc); err != nil {
		return err
	}

	expected := convertSystem(doc)

	var actual entity.System
	if err := m.pg.NewSelect().Model(&actual).Limit(1).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.AddMissing("systems", "system_record")

			return nil
		}

		return err
	}

	r.AddCompared("systems", 1)

	t := "systems"
	id := actual.ID

	r.CheckField(t, id, "Setup", expected.Setup, actual.Setup)
	r.CheckField(t, id, "Auth.Local.Enabled", expected.Authentication.Local.Enabled, actual.Authentication.Local.Enabled)
	r.CheckField(t, id, "Auth.SAML.Enabled", expected.Authentication.SAML.Enabled, actual.Authentication.SAML.Enabled)
	r.CheckField(t, id, "Auth.SAML.Idp.EntityID", expected.Authentication.SAML.Idp.EntityID, actual.Authentication.SAML.Idp.EntityID)
	r.CheckField(t, id, "Auth.SAML.Idp.Binding.Post", expected.Authentication.SAML.Idp.Binding.Post, actual.Authentication.SAML.Idp.Binding.Post)
	r.CheckField(t, id, "Auth.SAML.Idp.Binding.Redirect", expected.Authentication.SAML.Idp.Binding.Redirect, actual.Authentication.SAML.Idp.Binding.Redirect)
	r.CheckField(t, id, "Auth.SAML.Idp.Binding.Preferred", expected.Authentication.SAML.Idp.Binding.Preferred, actual.Authentication.SAML.Idp.Binding.Preferred)
	r.CheckStrings(t, id, "Auth.SAML.Idp.Certificates", expected.Authentication.SAML.Idp.Certificates, actual.Authentication.SAML.Idp.Certificates)
	r.CheckStringMap(t, id, "Auth.SAML.Idp.Mappings", expected.Authentication.SAML.Idp.Mappings, actual.Authentication.SAML.Idp.Mappings)
	r.CheckField(t, id, "Auth.SAML.Sp.SignAuthRequests", expected.Authentication.SAML.Sp.SignAuthRequests, actual.Authentication.SAML.Sp.SignAuthRequests)
	r.CheckFieldRedacted(t, id, "Auth.SAML.Sp.Certificate", expected.Authentication.SAML.Sp.Certificate, actual.Authentication.SAML.Sp.Certificate)
	r.CheckFieldRedacted(t, id, "Auth.SAML.Sp.PrivateKey", expected.Authentication.SAML.Sp.PrivateKey, actual.Authentication.SAML.Sp.PrivateKey)

	return nil
}
