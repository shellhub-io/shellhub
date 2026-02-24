package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(migration011Up, migration011Down)
}

func migration011Up(ctx context.Context, db *bun.DB) error {
	table := &struct {
		bun.BaseModel `bun:"table:systems"`

		ID                                    string            `bun:"id,type:uuid,pk"`
		Setup                                 bool              `bun:"setup,notnull,default:false"`
		AuthenticationLocalEnabled            bool              `bun:"authentication_local_enabled,notnull,default:true"`
		AuthenticationSamlEnabled             bool              `bun:"authentication_saml_enabled,notnull,default:false"`
		AuthenticationSamlIdpEntityID         string            `bun:"authentication_saml_idp_entity_id,type:text,nullzero"`
		AuthenticationSamlIdpBindingPost      string            `bun:"authentication_saml_idp_binding_post,type:text,nullzero"`
		AuthenticationSamlIdpBindingRedirect  string            `bun:"authentication_saml_idp_binding_redirect,type:text,nullzero"`
		AuthenticationSamlIdpBindingPreferred string            `bun:"authentication_saml_idp_binding_preferred,type:text,nullzero"`
		AuthenticationSamlIdpCertificates     []string          `bun:"authentication_saml_idp_certificates,array,nullzero"`
		AuthenticationSamlIdpMappings         map[string]string `bun:"authentication_saml_idp_mappings,type:jsonb,nullzero"`
		AuthenticationSamlSpSignAuthRequests  bool              `bun:"authentication_saml_sp_sign_auth_requests,notnull,default:false"`
		AuthenticationSamlSpCertificate       string            `bun:"authentication_saml_sp_certificate,type:text,nullzero"`
		AuthenticationSamlSpPrivateKey        string            `bun:"authentication_saml_sp_private_key,type:text,nullzero"`
	}{}

	if _, err := db.
		NewCreateTable().
		Model(table).
		IfNotExists().
		Exec(ctx); err != nil {
		log.WithError(err).Error("failed to apply migration 011")

		return err
	}

	return nil
}

func migration011Down(ctx context.Context, db *bun.DB) error {
	_, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS systems;`)
	if err != nil {
		log.WithError(err).Error("failed to revert migration 011")

		return err
	}

	return nil
}
