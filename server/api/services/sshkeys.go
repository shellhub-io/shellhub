package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"regexp"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"golang.org/x/crypto/ssh"
)

// PublicKeyFilterFields maps each filter field the public key list endpoint
// accepts to the set of operators valid for it.
var PublicKeyFilterFields = query.NewFieldConstraints(map[string][]string{
	"name":        {"contains", "eq", "ne"},
	"fingerprint": {"contains", "eq", "ne"},
})

// SSHKeysService owns a namespace's public keys and the rules restricting which devices and
// usernames each may reach.
type SSHKeysService interface {
	EvaluateKeyFilter(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error)
	EvaluateKeyUsername(ctx context.Context, key *models.PublicKey, username string) (bool, error)
	ListPublicKeys(ctx context.Context, req *requests.ListPublicKeys) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error)
	UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint, tenant string) error
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)
}

// Request is the template context for a key's username rule, so that a rule can be written
// in terms of the namespace it is evaluated in.
type Request struct {
	Namespace string
}

func (s *service) EvaluateKeyFilter(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error) {
	if len(key.Filter.TagIDs) > 0 {
		sc, err := BoundTo(key.TenantID)
		if err != nil {
			return false, err
		}
		d, err := s.store.DeviceResolve(ctx, sc, store.DeviceUIDResolver, dev.UID)
		if err != nil {
			return false, NewErrDeviceNotFound(models.UID(dev.UID), err)
		}

		dev.TagIDs = d.TagIDs
	}

	return key.Filter.Matches(&dev)
}

func (s *service) EvaluateKeyUsername(_ context.Context, key *models.PublicKey, username string) (bool, error) {
	if key.Username == "" {
		return true, nil
	}

	ok, err := regexp.MatchString(key.Username, username)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *service) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	sc, err := BoundTo(tenant)
	if err != nil {
		return nil, err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenant); err != nil {
		return nil, NewErrNamespaceNotFound(tenant, err)
	}

	publicKey, err := s.store.PublicKeyResolve(ctx, sc, store.PublicKeyFingerprintResolver, fingerprint)
	if err != nil {
		if errors.Is(err, store.ErrNoDocuments) {
			return nil, NewErrPublicKeyNotFound(fingerprint, err)
		}

		return nil, err
	}

	return publicKey, nil
}

func (s *service) CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error) {
	sc, err := BoundTo(tenant)
	if err != nil {
		return nil, err
	}

	tagIDs := []string{}
	if req.Filter.Tags != nil {
		tags, _, err := s.store.TagList(ctx, sc)
		if err != nil {
			return nil, NewErrTagEmpty(tenant, err)
		}

		for _, tagName := range req.Filter.Tags {
			found := false
			for _, tag := range tags {
				if tagName == tag.Name {
					tagIDs = append(tagIDs, tag.ID)
					found = true

					break
				}
			}

			if !found {
				return nil, NewErrTagNotFound(tagName, nil)
			}
		}
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(req.Data) //nolint:dogsled
	if err != nil {
		return nil, NewErrPublicKeyDataInvalid(req.Data, nil)
	}

	req.Fingerprint = ssh.FingerprintLegacyMD5(pubKey)

	returnedKey, err := s.store.PublicKeyResolve(ctx, sc, store.PublicKeyFingerprintResolver, req.Fingerprint)
	if err != nil && !errors.Is(err, store.ErrNoDocuments) {
		return nil, NewErrPublicKeyNotFound(req.Fingerprint, err)
	}

	if returnedKey != nil {
		return nil, NewErrPublicKeyDuplicated([]string{req.Fingerprint}, err)
	}

	model := models.PublicKey{
		Data:        ssh.MarshalAuthorizedKey(pubKey),
		Fingerprint: req.Fingerprint,
		CreatedAt:   clock.Now(),
		TenantID:    req.TenantID,
		PublicKeyFields: models.PublicKeyFields{
			Name:     req.Name,
			Username: req.Username,
			Filter: models.PublicKeyFilter{
				Hostname: req.Filter.Hostname,
				Taggable: models.Taggable{TagIDs: tagIDs, Tags: nil},
			},
		},
	}

	if _, err := s.store.PublicKeyCreate(ctx, &model); err != nil {
		return nil, err
	}

	return &responses.PublicKeyCreate{
		Data:        model.Data,
		Filter:      responses.PublicKeyFilter{Hostname: model.Filter.Hostname, Tags: req.Filter.Tags},
		Name:        model.Name,
		Username:    model.Username,
		TenantID:    model.TenantID,
		Fingerprint: model.Fingerprint,
	}, nil
}

func (s *service) ListPublicKeys(ctx context.Context, req *requests.ListPublicKeys) ([]models.PublicKey, int, error) {
	sc, err := BoundTo(req.TenantID)
	if err != nil {
		return nil, 0, err
	}

	return s.store.PublicKeyList(
		ctx,
		sc,
		s.store.Options().Match(&req.Filters),
		s.store.Options().Paginate(&req.Paginator),
	)
}

func (s *service) UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error) {
	sc, err := BoundTo(tenant)
	if err != nil {
		return nil, err
	}

	publicKey, err := s.store.PublicKeyResolve(ctx, sc, store.PublicKeyFingerprintResolver, fingerprint)
	if err != nil {
		return nil, NewErrPublicKeyNotFound(fingerprint, err)
	}

	tagIDs := []string{}
	if key.Filter.Tags != nil {
		tags, _, err := s.store.TagList(ctx, sc)
		if err != nil {
			return nil, NewErrTagEmpty(tenant, err)
		}

		for _, tagName := range key.Filter.Tags {
			found := false
			for _, tag := range tags {
				if tagName == tag.Name {
					tagIDs = append(tagIDs, tag.ID)
					found = true

					break
				}
			}

			if !found {
				return nil, NewErrTagNotFound(tagName, nil)
			}
		}
	}

	publicKey.Name = key.Name
	publicKey.Username = key.Username
	publicKey.Filter.Hostname = key.Filter.Hostname
	publicKey.Filter.TagIDs = tagIDs
	publicKey.Filter.Tags = nil

	if err := s.store.PublicKeyUpdate(ctx, publicKey); err != nil {
		return nil, err
	}

	return s.store.PublicKeyResolve(ctx, sc, store.PublicKeyFingerprintResolver, fingerprint)
}

func (s *service) DeletePublicKey(ctx context.Context, fingerprint, tenant string) error {
	sc, err := BoundTo(tenant)
	if err != nil {
		return err
	}

	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenant); err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	publicKey, err := s.store.PublicKeyResolve(ctx, sc, store.PublicKeyFingerprintResolver, fingerprint)
	if err != nil {
		return NewErrPublicKeyNotFound(fingerprint, err)
	}

	return s.store.PublicKeyDelete(ctx, publicKey)
}

func (s *service) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	pubKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey := &models.PrivateKey{
		Data: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
		Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
		CreatedAt:   clock.Now(),
	}

	if err := s.store.PrivateKeyCreate(ctx, privateKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}
