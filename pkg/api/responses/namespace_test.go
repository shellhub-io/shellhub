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

func TestNamespaceFromModelProjectsEveryField(t *testing.T) {
	createdAt := time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)
	settings := &models.NamespaceSettings{SessionRecord: true, Legacy: true}
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
			description: "projects the members of every namespace",
			namespaces: []models.Namespace{
				{
					Name: "first",
					Members: []models.Member{
						{ID: "human", Email: "human@test.com", Role: authorizer.RoleOwner},
					},
				},
				{
					Name: "second",
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
