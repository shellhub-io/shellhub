package responses_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceFromModel(t *testing.T) {
	cases := []struct {
		description string
		namespace   *models.Namespace
		expected    []responses.Member
	}{
		{
			description: "drops the service account",
			namespace: &models.Namespace{
				Members: []models.Member{
					{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
					{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService},
				},
			},
			expected: []responses.Member{
				{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
			},
		},
		{
			description: "drops a service account that does not carry the service role",
			namespace: &models.Namespace{
				Members: []models.Member{
					{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner, Type: models.UserTypeHuman},
					{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleObserver, Type: models.UserTypeService},
				},
			},
			expected: []responses.Member{
				{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
			},
		},
		{
			description: "keeps a member whose type is empty",
			namespace: &models.Namespace{
				Members: []models.Member{
					{ID: "legacy", Email: "legacy@test.com", Role: authorizer.RoleObserver},
				},
			},
			expected: []responses.Member{
				{ID: "legacy", Email: "legacy@test.com", Role: authorizer.RoleObserver},
			},
		},
		{
			description: "keeps a member whose type is human",
			namespace: &models.Namespace{
				Members: []models.Member{
					{ID: "human", Email: "human@test.com", Role: authorizer.RoleOperator, Type: models.UserTypeHuman},
				},
			},
			expected: []responses.Member{
				{ID: "human", Email: "human@test.com", Role: authorizer.RoleOperator},
			},
		},
		{
			description: "projects an empty list when every member is a service account",
			namespace: &models.Namespace{
				Members: []models.Member{
					{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService},
				},
			},
			expected: []responses.Member{},
		},
		{
			description: "projects an empty list when the namespace has no members",
			namespace:   &models.Namespace{},
			expected:    []responses.Member{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			res := responses.NamespaceFromModel(tc.namespace)

			require.NotNil(t, res)
			assert.Equal(t, tc.expected, res.Members)
		})
	}
}

func TestNamespaceFromModelNil(t *testing.T) {
	assert.Nil(t, responses.NamespaceFromModel(nil))
}

func TestNamespaceFromModelNeverMarshalsMembersAsNull(t *testing.T) {
	data, err := json.Marshal(responses.NamespaceFromModel(&models.Namespace{}))
	require.NoError(t, err)

	decoded := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.JSONEq(t, "[]", string(decoded["members"]))
}

func TestNamespaceFromModelDoesNotMutateSource(t *testing.T) {
	namespace := &models.Namespace{
		Members: []models.Member{
			{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
			{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService},
		},
	}

	responses.NamespaceFromModel(namespace)

	assert.Equal(
		t,
		[]models.Member{
			{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
			{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService},
		},
		namespace.Members,
	)
}

func TestNamespaceFromModelProjectsEveryField(t *testing.T) {
	createdAt := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	settings := &models.NamespaceSettings{SessionRecord: true, SSHAccessMode: "identity"}
	billing := &models.Billing{CustomerID: "cus_test"}

	res := responses.NamespaceFromModel(&models.Namespace{
		Name:                 "namespace",
		Owner:                "000000000000000000000000",
		TenantID:             "00000000-0000-4000-0000-000000000000",
		Members:              []models.Member{{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner}},
		Settings:             settings,
		MaxDevices:           10,
		CreatedAt:            createdAt,
		Billing:              billing,
		Type:                 models.TypeTeam,
		DevicesAcceptedCount: 1,
		DevicesPendingCount:  2,
		DevicesRejectedCount: 3,
		DevicesRemovedCount:  4,
	})

	assert.Equal(
		t,
		&responses.Namespace{
			Name:                 "namespace",
			Owner:                "000000000000000000000000",
			TenantID:             "00000000-0000-4000-0000-000000000000",
			Members:              []responses.Member{{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner}},
			Settings:             settings,
			MaxDevices:           10,
			CreatedAt:            createdAt,
			Billing:              billing,
			Type:                 models.TypeTeam,
			DevicesAcceptedCount: 1,
			DevicesPendingCount:  2,
			DevicesRejectedCount: 3,
			DevicesRemovedCount:  4,
		},
		res,
	)
}

func TestNamespacesFromModel(t *testing.T) {
	cases := []struct {
		description string
		namespaces  []models.Namespace
		expected    []responses.Namespace
	}{
		{
			description: "projects an empty list when there are no namespaces",
			namespaces:  []models.Namespace{},
			expected:    []responses.Namespace{},
		},
		{
			description: "projects an empty list when the source is nil",
			namespaces:  nil,
			expected:    []responses.Namespace{},
		},
		{
			description: "drops the service account from each namespace",
			namespaces: []models.Namespace{
				{
					Name: "first",
					Members: []models.Member{
						{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
						{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService},
					},
				},
				{
					Name:    "second",
					Members: []models.Member{{ID: "bot", Email: "bot@test.com", Role: authorizer.RoleService, Type: models.UserTypeService}},
				},
			},
			expected: []responses.Namespace{
				{
					Name:    "first",
					Members: []responses.Member{{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner}},
				},
				{
					Name:    "second",
					Members: []responses.Member{},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, responses.NamespacesFromModel(tc.namespaces))
		})
	}
}
