package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"golang.org/x/crypto/ssh"
)

type ServiceAccountService interface {
	// CreateServiceAccount creates a service account in the namespace: a service-typed
	// user, a membership with the service role, and an SSH identity for the given public
	// key, all atomically. The account never signs in and is not an API principal.
	CreateServiceAccount(ctx context.Context, req *requests.ServiceAccountCreate) (*models.ServiceAccount, error)

	// ListServiceAccounts returns the namespace's service accounts with their identities.
	ListServiceAccounts(ctx context.Context, req *requests.ServiceAccountList) ([]models.ServiceAccount, error)

	// DeleteServiceAccount removes a service account. Deleting the account cascades to its
	// membership and every SSH identity it holds.
	DeleteServiceAccount(ctx context.Context, req *requests.ServiceAccountDelete) error
}

func (s *service) CreateServiceAccount(ctx context.Context, req *requests.ServiceAccountCreate) (*models.ServiceAccount, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	// Parse the key up front so an invalid key fails before opening the transaction.
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.Data)) //nolint:dogsled
	if err != nil {
		return nil, NewErrSSHIdentityInvalid(req.Data, err)
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)
	data := ssh.MarshalAuthorizedKey(pubKey)

	// Service accounts carry no console credentials. Their identifiers are synthetic and
	// never shown; the password digest is a locked sentinel and login rejects the type
	// anyway (see AuthLocalUser). auth_methods is empty so no method can authenticate them.
	suffix := uuid.Generate()
	user := &models.User{
		Type:          models.UserTypeService,
		Origin:        models.UserOriginLocal,
		Status:        models.UserStatusConfirmed,
		MaxNamespaces: 0,
		CreatedAt:     clock.Now(),
		UserData: models.UserData{
			Name:     req.Name,
			Username: "svc-" + suffix,
			Email:    "svc-" + suffix + "@service.local",
		},
		Password: models.UserPassword{Hash: "!"},
		Preferences: models.UserPreferences{
			AuthMethods: []models.UserAuthMethod{},
		},
	}

	account := &models.ServiceAccount{}
	err = s.store.WithTransaction(ctx, func(ctx context.Context) error {
		id, err := s.store.UserCreate(ctx, user)
		if err != nil {
			return err
		}

		member := &models.Member{ID: id, AddedAt: clock.Now(), Role: authorizer.RoleService}
		if err := s.admitMember(ctx, req.TenantID, member, nil); err != nil {
			return err
		}

		identity, err := s.enrollSSHIdentity(ctx, id, req.TenantID, fingerprint, data, req.Name)
		if err != nil {
			return err
		}

		account.ID = id
		account.Name = user.Name
		account.CreatedAt = user.CreatedAt
		account.Identities = []models.SSHIdentity{*identity}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *service) ListServiceAccounts(ctx context.Context, req *requests.ServiceAccountList) ([]models.ServiceAccount, error) {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID); err != nil {
		return nil, NewErrNamespaceNotFound(req.TenantID, err)
	}

	accounts, _, err := s.store.ServiceAccountList(ctx, req.TenantID)
	if err != nil {
		return nil, err
	}

	identities, _, err := s.store.SSHIdentityList(ctx, s.store.Options().InNamespace(req.TenantID))
	if err != nil {
		return nil, err
	}

	byUser := make(map[string][]models.SSHIdentity, len(accounts))
	for _, identity := range identities {
		byUser[identity.PrincipalID] = append(byUser[identity.PrincipalID], identity)
	}

	for i := range accounts {
		accounts[i].Identities = byUser[accounts[i].ID]
	}

	return accounts, nil
}

func (s *service) DeleteServiceAccount(ctx context.Context, req *requests.ServiceAccountDelete) error {
	user, err := s.store.UserResolve(ctx, store.UserIDResolver, req.ID)
	if err != nil || user.Type != models.UserTypeService {
		return NewErrServiceAccountNotFound(req.ID, err)
	}

	// Confirm the service account belongs to the requesting namespace before deleting it,
	// so a caller can't remove one from another namespace by id.
	namespace, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, req.TenantID)
	if err != nil {
		return NewErrNamespaceNotFound(req.TenantID, err)
	}

	if _, ok := namespace.FindMember(req.ID); !ok {
		return NewErrServiceAccountNotFound(req.ID, nil)
	}

	// Deleting the user cascades to its membership and SSH identities (FK ON DELETE CASCADE).
	return s.store.UserDelete(ctx, &models.User{ID: req.ID})
}
