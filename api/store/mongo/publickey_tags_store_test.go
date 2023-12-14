package mongo

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyAddTag(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyAddTag(context.TODO(), tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyRemoveTag(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyRemoveTag(context.TODO(), tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyUpdateTags(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tags        []string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyUpdateTags(context.TODO(), tc.tenant, tc.fingerprint, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyRenameTag(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "nonexistent",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyRenameTag(context.TODO(), tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyDeleteTag(t *testing.T) {
	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to tenant",
			tenant:      "nonexistent",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyDeleteTag(context.TODO(), tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyGetTags(t *testing.T) {
	type Expected struct {
		tags []string
		len  int
		err  error
	}

	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when tags list is greater than 1",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			tags, count, err := mongostore.PublicKeyGetTags(context.TODO(), tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
