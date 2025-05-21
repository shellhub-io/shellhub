package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
)

func TestPrivateKeyCreate(t *testing.T) {
	cases := []struct {
		description string
		priKey      *models.PrivateKey
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			priKey: &models.PrivateKey{
				Data:        []byte("test"),
				Fingerprint: "fingerprint",
				CreatedAt:   time.Now(),
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.PrivateKeyCreate(ctx, tc.priKey)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPrivateKeyGet(t *testing.T) {
	type Expected struct {
		privKey *models.PrivateKey
		err     error
	}

	cases := []struct {
		description string
		fingerprint string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when private key is not found",
			fingerprint: "nonexistent",
			fixtures:    []string{fixturePrivateKeys},
			expected: Expected{
				privKey: nil,
				err:     store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when private key is found",
			fingerprint: "fingerprint",
			fixtures:    []string{fixturePrivateKeys},
			expected: Expected{
				privKey: &models.PrivateKey{
					Data:        []byte("test"),
					Fingerprint: "fingerprint",
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			privKey, err := s.PrivateKeyGet(ctx, tc.fingerprint)
			assert.Equal(t, tc.expected, Expected{privKey: privKey, err: err})
		})
	}
}
