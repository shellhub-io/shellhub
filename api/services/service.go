package services

import (
	"context"
	"crypto/rsa"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

type Service interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filter string, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID, tenant, ownerID string) error
	RenameDevice(ctx context.Context, uid models.UID, name, tenant, ownerID string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status, tenant, ownerID string) error

	UpdateDataUser(ctx context.Context, data *models.User, id string) ([]validator.InvalidField, error)
	UpdatePasswordUser(ctx context.Context, currentPassword, newPassword, id string) error

	EvaluateKeyHostname(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error)
	ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, key *models.PublicKey, tenant string) error
	UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key *models.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint, tenant string) error
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)

	ListSessions(ctx context.Context, pagination paginator.Query) ([]models.Session, int, error)
	GetSession(ctx context.Context, uid models.UID) (*models.Session, error)
	CreateSession(ctx context.Context, session models.Session) (*models.Session, error)
	DeactivateSession(ctx context.Context, uid models.UID) error
	SetSessionAuthenticated(ctx context.Context, uid models.UID, authenticated bool) error

	ListNamespaces(ctx context.Context, pagination paginator.Query, filterB64 string, export bool) ([]models.Namespace, int, error)
	CreateNamespace(ctx context.Context, namespace *models.Namespace, ownerUsername string) (*models.Namespace, error)
	GetNamespace(ctx context.Context, tenantID string) (*models.Namespace, error)
	DeleteNamespace(ctx context.Context, tenantID, ownerUsername string) error
	EditNamespace(ctx context.Context, tenantID, name, ownerUsername string) (*models.Namespace, error)
	AddNamespaceUser(ctx context.Context, tenantID, username, ownerUsername string) (*models.Namespace, error)
	RemoveNamespaceUser(ctx context.Context, tenantID, username, ownerUsername string) (*models.Namespace, error)
	ListMembers(ctx context.Context, tenantID string) ([]models.Member, error)
	EditSessionRecordStatus(ctx context.Context, status bool, tenant, ownerID string) error
	GetSessionRecord(ctx context.Context, tenant string) (bool, error)

	AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error)
	AuthGetToken(ctx context.Context, tenant string) (*models.UserAuthResponse, error)
	AuthPublicKey(ctx context.Context, req *models.PublicKeyAuthRequest) (*models.PublicKeyAuthResponse, error)
	AuthSwapToken(ctx context.Context, ID, tenant string) (*models.UserAuthResponse, error)
	AuthUserInfo(ctx context.Context, username, tenant, token string) (*models.UserAuthResponse, error)
	PublicKey() *rsa.PublicKey

	GetStats(ctx context.Context) (*models.Stats, error)
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = LoadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey}
}
