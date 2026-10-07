package services

import (
	"context"
	"errors"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	log "github.com/sirupsen/logrus"
)

// devNamespace / devTenantID are the well-known fixtures used across the development stack (e.g.
// the built-in dev agent connects to the "dev" namespace on this tenant). In development, a setup
// that keeps the default "dev" name binds to this tenant so those fixtures keep working; renaming
// the namespace opts out and a fresh tenant is generated (to exercise the normal flow).
const (
	devNamespace = "dev"
	devTenantID  = "00000000-0000-4000-0000-000000000000"
)

// SetupService creates the first user and namespace on a fresh instance. It refuses once an
// instance is set up, so it cannot be used to add a second administrator.
type SetupService interface {
	// Setup creates the first user, as the instance's administrator, with their namespace, binds
	// the instance to that namespace and signs the user in. It returns ErrSetupCompleted once the
	// instance is set up; ErrUserInvalid or ErrUserPasswordInvalid for a request that fails
	// validation; ErrUserDuplicated or ErrUserUnhandledDuplicate when the user conflicts with an
	// existing one; ErrNamespaceDuplicated or ErrNamespaceCreateStore when the namespace cannot be
	// created, after removing the user again, or ErrUserDelete when that removal fails; and the
	// store's error when the system row cannot be read or written or the user cannot be stored.
	// A failed write of the system row leaves the user and namespace in place. When no token can
	// be issued for the new user, it returns an empty response and no error.
	Setup(ctx context.Context, req requests.Setup) (*models.UserAuthResponse, error)
}

func (s *service) Setup(ctx context.Context, req requests.Setup) (*models.UserAuthResponse, error) {
	system, err := s.store.SystemGet(ctx)
	if err != nil {
		return nil, err
	}

	if system.Setup {
		return nil, NewErrSetupCompleted()
	}

	data := models.UserData{
		Name:          req.Name,
		Email:         req.Email,
		Username:      req.Username,
		RecoveryEmail: "",
	}

	if ok, err := s.validator.Struct(data); !ok || err != nil {
		return nil, NewErrUserInvalid(nil, err)
	}

	password, err := models.HashUserPassword(req.Password)
	if err != nil {
		return nil, NewErrUserPasswordInvalid(err)
	}

	if ok, err := s.validator.Struct(password); !ok || err != nil {
		return nil, NewErrUserPasswordInvalid(err)
	}

	user := &models.User{
		Origin:        models.UserOriginLocal,
		UserData:      data,
		Password:      password,
		Status:        models.UserStatusConfirmed,
		CreatedAt:     clock.Now(),
		MaxNamespaces: -1,
		Preferences: models.UserPreferences{
			AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
		},
		Admin: true,
	}

	insertedID, err := s.store.UserCreate(ctx, user)
	if err != nil {
		if errors.Is(err, store.ErrDuplicate) {
			if field, ok := store.DuplicatedField(err); ok {
				return nil, NewErrUserDuplicated([]string{field}, err)
			}

			return nil, NewErrUserUnhandledDuplicate()
		}

		return nil, err
	}

	namespaceName := strings.ToLower(req.Namespace)

	tenantID := uuid.Generate()
	if envs.IsDevelopment() && namespaceName == devNamespace {
		tenantID = devTenantID
	}

	namespace := &models.Namespace{
		Name:       namespaceName,
		TenantID:   tenantID,
		MaxDevices: -1,
		Owner:      insertedID,
		Type:       models.TypePersonal,
		Members: []models.Member{
			{
				ID:      insertedID,
				Role:    authorizer.RoleOwner,
				AddedAt: clock.Now(),
			},
		},
		CreatedAt: clock.Now(),
		Settings: &models.NamespaceSettings{
			SessionRecord:          false,
			ConnectionAnnouncement: envs.AnnouncementFor(envs.CurrentEdition()),
		},
	}

	if _, err = s.store.NamespaceCreate(ctx, namespace); err != nil {
		user.ID = insertedID
		if err := s.store.UserDelete(ctx, user); err != nil {
			return nil, NewErrUserDelete(err)
		}

		if errors.Is(err, store.ErrDuplicate) {
			return nil, NewErrNamespaceDuplicated(err)
		}

		return nil, NewErrNamespaceCreateStore(err)
	}

	system.Setup = true
	system.InstanceTenantID = namespace.TenantID
	if err := s.store.SystemSet(ctx, system); err != nil {
		return nil, err
	}

	if err := s.cache.Delete(ctx, cache.SystemKey); err != nil {
		log.WithError(err).Warn("failed to evict the cached system row after setup")
	}

	res, err := s.CreateUserToken(ctx, &requests.CreateUserToken{UserID: insertedID, TenantID: namespace.TenantID})
	if err != nil {
		log.WithError(err).Warn("setup completed but failed to issue an auto-login token")

		return &models.UserAuthResponse{}, nil
	}

	return res, nil
}
