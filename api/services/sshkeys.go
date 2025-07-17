package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"regexp"
	"slices"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

type SSHKeysService interface {
	EvaluateKeyFilter(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error)
	EvaluateKeyUsername(ctx context.Context, key *models.PublicKey, username string) (bool, error)
	ListPublicKeys(ctx context.Context, paginator query.Paginator) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error)
	UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint, tenant string) error
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)
}

type Request struct {
	Namespace string
}

func (s *service) EvaluateKeyFilter(_ context.Context, key *models.PublicKey, dev models.Device) (bool, error) {
	if key.Filter.Hostname != "" {
		ok, err := regexp.MatchString(key.Filter.Hostname, dev.Name)
		if err != nil {
			return false, err
		}

		return ok, nil
	} else if len(key.Filter.Tags) > 0 {
		tagNames := make([]string, 0)
		for _, tag := range key.Filter.Tags {
			tagNames = append(tagNames, tag.Name)
		}

		for _, tag := range dev.Tags {
			if slices.Contains(tagNames, tag.Name) {
				return true, nil
			}
		}

		return false, nil
	}

	return true, nil
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
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenant); err != nil {
		return nil, NewErrNamespaceNotFound(tenant, err)
	}

	return s.store.PublicKeyGet(ctx, fingerprint, tenant)
}

func (s *service) CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error) {
	tagIDs := make([]string, 0)
	// Checks if public key filter type is Tags.
	// If it is, checks if there are, at least, one tag on the public key filter and if the all tags exist on database.
	if req.Filter.Tags != nil {
		tags, _, err := s.store.TagList(ctx, tenant, query.Paginator{Page: 1, PerPage: 9999}, query.Filters{}, query.Sorter{})
		if err != nil {
			return nil, err
		}

		tagNameToID := make(map[string]string, len(tags))
		for _, tag := range tags {
			tagNameToID[tag.Name] = tag.ID
		}

		for _, tagName := range req.Filter.Tags {
			id, ok := tagNameToID[tagName]
			if !ok {
				return nil, NewErrTagNotFound(tagName, nil)
			}

			tagIDs = append(tagIDs, id)
		}
	}

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(req.Data) //nolint:dogsled
	if err != nil {
		return nil, NewErrPublicKeyDataInvalid(req.Data, nil)
	}

	req.Fingerprint = ssh.FingerprintLegacyMD5(pubKey)

	returnedKey, err := s.store.PublicKeyGet(ctx, req.Fingerprint, tenant)
	if err != nil && err != store.ErrNoDocuments {
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
				Taggable: models.Taggable{TagsID: tagIDs},
			},
		},
	}

	err = s.store.PublicKeyCreate(ctx, &model)
	if err != nil {
		return nil, err
	}

	// Response filter conversion: internal storage uses tag IDs, external API uses tag names
	responseFilter := responses.PublicKeyFilter{Hostname: model.Filter.Hostname, Tags: req.Filter.Tags}

	return &responses.PublicKeyCreate{
		Data:        model.Data,
		Filter:      responseFilter,
		Name:        model.Name,
		Username:    model.Username,
		TenantID:    model.TenantID,
		Fingerprint: model.Fingerprint,
	}, nil
}

func (s *service) ListPublicKeys(ctx context.Context, paginator query.Paginator) ([]models.PublicKey, int, error) {
	return s.store.PublicKeyList(ctx, paginator)
}

func (s *service) UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error) {
	tagIDs := make([]string, 0)
	// Checks if public key filter type is Tags.
	// If it is, checks if there are, at least, one tag on the public key filter and if the all tags exist on database.
	if key.Filter.Tags != nil {
		tags, _, err := s.store.TagList(ctx, tenant, query.Paginator{Page: 1, PerPage: 9999}, query.Filters{}, query.Sorter{})
		if err != nil {
			return nil, err
		}

		tagNameToID := make(map[string]string, len(tags))
		for _, tag := range tags {
			tagNameToID[tag.Name] = tag.ID
		}

		for _, tagName := range key.Filter.Tags {
			id, ok := tagNameToID[tagName]
			if !ok {
				return nil, NewErrTagNotFound(tagName, nil)
			}

			tagIDs = append(tagIDs, id)
		}
	}

	model := models.PublicKeyUpdate{
		PublicKeyFields: models.PublicKeyFields{
			Name:     key.Name,
			Username: key.Username,
			Filter: models.PublicKeyFilter{
				Hostname: key.Filter.Hostname,
				Taggable: models.Taggable{TagsID: tagIDs},
			},
		},
	}

	return s.store.PublicKeyUpdate(ctx, fingerprint, tenant, &model)
}

func (s *service) DeletePublicKey(ctx context.Context, fingerprint, tenant string) error {
	if _, err := s.store.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, tenant); err != nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	if _, err := s.store.PublicKeyGet(ctx, fingerprint, tenant); err != nil {
		return NewErrPublicKeyNotFound(fingerprint, err)
	}

	return s.store.PublicKeyDelete(ctx, fingerprint, tenant)
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
