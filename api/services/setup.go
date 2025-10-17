package services

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"os"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

const PrivateKeyPath = "/var/run/secrets/api_private_key"

type SetupService interface {
	Setup(ctx context.Context, req requests.Setup) error
	SetupVerify(ctx context.Context, sign string) error
}

func (s *service) Setup(ctx context.Context, req requests.Setup) error {
	if system, err := s.store.SystemGet(ctx); err != nil || system.Setup {
		return NewErrSetupForbidden(err)
	}

	data := models.UserData{
		Name:          req.Name,
		Email:         req.Email,
		Username:      req.Username,
		RecoveryEmail: "",
	}

	if ok, err := s.validator.Struct(data); !ok || err != nil {
		return NewErrUserInvalid(nil, err)
	}

	password, err := models.HashUserPassword(req.Password)
	if err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return NewErrUserPasswordInvalid(err)
	}

	// Verificar se é o primeiro usuário (apenas em Community/Enterprise)
	var superAdmin bool
	if !envs.IsCloud() {
		firstUser, err := s.isFirstUser(ctx)
		if err != nil {
			return err
		}
		superAdmin = firstUser
	}
	// Na Cloud, superAdmin permanece false (valor padrão)

	user := &models.User{
		Origin:   models.UserOriginLocal,
		UserData: data,
		Password: password,
		// NOTE: user's created from the setup screen doesn't need to be confirmed.
		Status:        models.UserStatusConfirmed,
		CreatedAt:     clock.Now(),
		MaxNamespaces: -1,
		Preferences: models.UserPreferences{
			AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
		},
		SuperAdmin: superAdmin,
	}

	insertedID, err := s.store.UserCreate(ctx, user)
	if err != nil {
		return NewErrUserDuplicated([]string{req.Username}, err)
	}

	namespace := &models.Namespace{
		Name:       req.Username,
		TenantID:   uuid.Generate(),
		MaxDevices: -1,
		Owner:      insertedID,
		Type:       models.TypePersonal,
		Members: []models.Member{
			{
				ID:      insertedID,
				Role:    authorizer.RoleOwner,
				Status:  models.MemberStatusAccepted,
				AddedAt: clock.Now(),
			},
		},
		CreatedAt: clock.Now(),
		Settings: &models.NamespaceSettings{
			SessionRecord:          false,
			ConnectionAnnouncement: models.DefaultAnnouncementMessage,
		},
	}

	if _, err = s.store.NamespaceCreate(ctx, namespace); err != nil {
		if err := s.store.UserDelete(ctx, insertedID); err != nil {
			return NewErrUserDelete(err)
		}

		return NewErrNamespaceDuplicated(err)
	}

	if err := s.store.SystemSet(ctx, "setup", true); err != nil { //nolint:revive
		return err
	}

	return nil
}

func (s *service) SetupVerify(_ context.Context, sign string) error {
	privKeyData, err := os.ReadFile(PrivateKeyPath)
	if err != nil {
		return err
	}

	privKeyPem, _ := pem.Decode(privKeyData)
	privKey, err := x509.ParsePKCS8PrivateKey(privKeyPem.Bytes)
	if err != nil {
		return err
	}

	const msgString = "shellhub"

	msgHash := sha256.New()
	_, err = msgHash.Write([]byte(msgString))
	if err != nil {
		return err
	}

	signed, err := rsa.SignPKCS1v15(rand.Reader, privKey.(*rsa.PrivateKey), crypto.SHA256, msgHash.Sum(nil))
	if err != nil {
		return err
	}

	sumSigned := sha256.Sum256(signed)

	if sign != hex.EncodeToString(sumSigned[:]) {
		return NewErrSetupForbidden(nil)
	}

	return nil
}
