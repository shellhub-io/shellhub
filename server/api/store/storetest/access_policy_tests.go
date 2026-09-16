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
	s.CreateMembership(t, tenantID, ownerID, "operator")
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

// TestAccessPolicySubjectMustMatchSomeone verifies the schema refuses a subject that could never
// match anyone, so an inert rule cannot be stored whatever reaches the store.
func (s *Suite) TestAccessPolicySubjectMustMatchSomeone(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	memberID := s.CreateUser(t)
	strangerID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, memberID, "operator")

	policy := func(subject models.PolicySubject) *models.AccessPolicy {
		return &models.AccessPolicy{
			TenantID: tenantID,
			Name:     "rule",
			Subject:  subject,
			Logins:   []string{"root"},
			SourceIP: []string{},
			Action:   models.PolicyActionDeny,
		}
	}

	_, err := st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectUser, Value: memberID}))
	require.NoError(t, err, "a member of the namespace is a subject that can match")

	_, err = st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectUser, Value: strangerID}))
	require.Error(t, err, "a user who is not a member of this namespace matches nobody")

	_, err = st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectRole, Value: "operator"}))
	require.NoError(t, err, "a role the namespace defines is a subject that can match")

	_, err = st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectRole, Value: "superadmin"}))
	require.Error(t, err, "a role nothing defines matches nobody")

	_, err = st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectAllMembers}))
	require.NoError(t, err, "every member is a subject that can match")

	id, err := st.AccessPolicyCreate(ctx, policy(models.PolicySubject{Type: models.PolicySubjectAllMembers, Value: memberID}))
	require.NoError(t, err)

	stored, err := st.AccessPolicyResolve(ctx, scope.MustBounded(tenantID), store.AccessPolicyIDResolver, id)
	require.NoError(t, err)
	assert.Empty(t, stored.Subject.Value, "every member matches by type, so a value on it has nowhere to be stored")
}

// TestAccessPolicyGoesWithTheMembershipItNames verifies that removing a member takes the rules
// naming them, so a rule never outlives the person it was written about. Applies to both actions:
// the deny is the one that would otherwise come back to life if they rejoined.
func (s *Suite) TestAccessPolicyGoesWithTheMembershipItNames(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	leaverID := s.CreateUser(t)
	stayerID := s.CreateUser(t)
	s.CreateMembership(t, tenantID, leaverID, "operator")
	s.CreateMembership(t, tenantID, stayerID, "operator")

	for _, subject := range []string{leaverID, stayerID} {
		_, err := st.AccessPolicyCreate(ctx, &models.AccessPolicy{
			TenantID: tenantID,
			Name:     "rule-" + subject,
			Subject:  models.PolicySubject{Type: models.PolicySubjectUser, Value: subject},
			Logins:   []string{"root"},
			SourceIP: []string{},
			Action:   models.PolicyActionDeny,
		})
		require.NoError(t, err)
	}

	subjects := func() []string {
		policies, _, err := st.AccessPolicyList(ctx, scope.MustBounded(tenantID))
		require.NoError(t, err)

		named := make([]string, 0, len(policies))
		for _, policy := range policies {
			if policy.Subject.Value == leaverID || policy.Subject.Value == stayerID {
				named = append(named, policy.Subject.Value)
			}
		}

		return named
	}

	require.ElementsMatch(t, []string{leaverID, stayerID}, subjects())

	require.NoError(t, st.NamespaceDeleteMembership(ctx, scope.MustBounded(tenantID), &models.Member{ID: leaverID}))

	assert.Equal(t, []string{stayerID}, subjects(), "only the rule naming the member who stayed survives")
}
