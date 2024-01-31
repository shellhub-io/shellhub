package mongo_test

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyPushTag(t *testing.T) {
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
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tag:         "new-tag",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "new-tag",
			fixtures:    []string{fixturePublicKeys},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := s.PublicKeyPushTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyPullTag(t *testing.T) {
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
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tag",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			err := s.PublicKeyPullTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeySetTags(t *testing.T) {
	type Expected struct {
		matchedCount int64
		updatedCount int64
		err          error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		tags        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds when tags public key is found and tags are equal to current public key tags",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"tag-1"},
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "succeeds when tags public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tags:        []string{"new-tag"},
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 1,
				err:          nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			matchedCount, updatedCount, err := s.PublicKeySetTags(ctx, tc.tenant, tc.fingerprint, tc.tags)
			assert.Equal(t, tc.expected, Expected{matchedCount, updatedCount, err})
		})
	}
}

func TestPublicKeyBulkRenameTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when public key is not found due to tenant",
			tenant:      "",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "",
			newTag:      "edited-tag",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "edited-tag",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			count, err := s.PublicKeyBulkRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestPublicKeyBulkDeleteTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when public key is not found due to tenant",
			tenant:      "",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when public key is not found due to tag",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "succeeds when public key is found",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				count: 1,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			count, err := s.PublicKeyBulkDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
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
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, db.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, db.Reset())
			})

			tags, count, err := s.PublicKeyGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
