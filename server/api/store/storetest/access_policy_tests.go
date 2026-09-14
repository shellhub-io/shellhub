package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessPolicyTagFilterRoundTrip verifies the device selector survives the join table: tags go
// in as ids and come back as whole tags.
func (s *Suite) TestAccessPolicyTagFilterRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	ownerID := s.CreateUser(t)
	tagID := s.CreateTag(t, WithTagTenant(tenantID), WithTagName("production"))

	id, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
		TenantID: tenantID,
		Name:     "ops on production",
		Subject:  models.PolicySubject{Type: models.PolicySubjectUser, Value: ownerID},
		Filter:   models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{tagID}}},
		Logins:   []string{"root"},
		SourceIP: []string{},
		Action:   models.PolicyActionAllow,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	sc := scope.MustBounded(tenantID)

	policy, err := st.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, id)
	require.NoError(t, err)
	require.NotNil(t, policy)

	assert.Equal(t, "ops on production", policy.Name)
	assert.Equal(t, models.PolicySubjectUser, policy.Subject.Type)
	assert.Equal(t, ownerID, policy.Subject.Value)
	assert.Equal(t, []string{"root"}, policy.Logins)
	assert.Equal(t, models.PolicyActionAllow, policy.Action)

	require.Len(t, policy.Filter.Tags, 1)
	assert.Equal(t, "production", policy.Filter.Tags[0].Name)
	assert.Equal(t, []string{tagID}, policy.Filter.TagIDs)
}

// TestAccessPolicyUpdateReplacesTheTagFilter verifies an update swaps the selector instead of
// adding to it.
func (s *Suite) TestAccessPolicyUpdateReplacesTheTagFilter(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	staging := s.CreateTag(t, WithTagTenant(tenantID), WithTagName("staging"))
	production := s.CreateTag(t, WithTagTenant(tenantID), WithTagName("production"))

	id, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
		TenantID: tenantID,
		Name:     "everywhere",
		Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
		Filter:   models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{staging, production}}},
		Logins:   []string{"*"},
		SourceIP: []string{},
		Action:   models.PolicyActionAllow,
	})
	require.NoError(t, err)

	sc := scope.MustBounded(tenantID)

	require.NoError(t, st.AccessPolicyUpdate(ctx, &models.AccessPolicy{
		ID:       id,
		TenantID: tenantID,
		Name:     "production only",
		Subject:  models.PolicySubject{Type: models.PolicySubjectAllMembers},
		Filter:   models.PublicKeyFilter{Taggable: models.Taggable{TagIDs: []string{production}}},
		Logins:   []string{"*"},
		SourceIP: []string{},
		Action:   models.PolicyActionDeny,
	}))

	policy, err := st.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, id)
	require.NoError(t, err)

	assert.Equal(t, "production only", policy.Name)
	assert.Equal(t, models.PolicyActionDeny, policy.Action)
	require.Len(t, policy.Filter.Tags, 1)
	assert.Equal(t, "production", policy.Filter.Tags[0].Name)

	require.NoError(t, st.AccessPolicyDelete(ctx, policy))

	_, err = st.AccessPolicyResolve(ctx, sc, store.AccessPolicyIDResolver, id)
	assert.Error(t, err)
}
