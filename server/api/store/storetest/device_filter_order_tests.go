package storetest

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceListFilterEvaluationOrder pins how a filter joins its conditions: left to right, an
// operator applying to the condition after it, OR before any operator, a trailing operator ignored,
// AND binding tighter than OR, and a group evaluated as one bracketed condition. No filter reaches
// another namespace's devices, and an invalid filter is an error.
func (s *Suite) TestDeviceListFilterEvaluationOrder(t *testing.T) {
	ctx := context.Background()
	st := s.provider.Store()

	require.NoError(t, s.provider.CleanDatabase(t))

	tenantID := s.CreateNamespace(t)
	prod := s.CreateTag(t, WithTagName("prod"), WithTagTenant(tenantID))

	tagged := func(tenantID, tagID, name string) string {
		uid := string(s.CreateDevice(t, WithDeviceName(name), WithTenantID(tenantID)))
		require.NoError(t, st.TagPushToTarget(ctx, tagID, store.TagTargetDevice, uid))

		return uid
	}

	s.CreateDevice(t, WithDeviceName("web-untagged"), WithTenantID(tenantID))
	tagged(tenantID, prod, "db-tagged")
	tagged(tenantID, prod, "web-tagged")
	custom := tagged(tenantID, prod, "db-custom")
	require.NoError(t, st.DeviceSetCustomField(ctx, scope.MustBounded(tenantID), custom, "role", "web"))

	otherTenantID := s.CreateNamespace(t)
	tagged(otherTenantID, s.CreateTag(t, WithTagName("prod"), WithTagTenant(otherTenantID)), "web-other")

	name := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "web"}}
	customFields := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "custom_fields", Operator: "contains", Value: "web"}}
	tag := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "tags.name", Operator: "contains", Value: []any{"prod"}}}
	and := query.Filter{Type: query.FilterTypeOperator, Params: &query.FilterOperator{Name: "and"}}
	or := query.Filter{Type: query.FilterTypeOperator, Params: &query.FilterOperator{Name: "or"}}
	group := func(nodes ...query.Filter) query.Filter {
		return query.Filter{Type: query.FilterTypeGroup, Params: query.FilterGroup(nodes)}
	}

	list := func(filters []query.Filter) ([]string, error) {
		devices, _, err := st.DeviceList(ctx, scope.MustBounded(tenantID), store.DeviceAcceptableIfNotAccepted, st.Options().Match(&query.Filters{Data: filters}))

		names := make([]string, len(devices))
		for i, device := range devices {
			names[i] = device.Name
		}

		return names, err
	}

	cases := []struct {
		description string
		filters     []query.Filter
		want        []string
	}{
		{
			description: "a condition with no operator before it is ORed",
			filters:     []query.Filter{name, tag},
			want:        []string{"web-untagged", "db-tagged", "web-tagged", "db-custom"},
		},
		{
			description: "leading or operators keep every condition ORed",
			filters:     []query.Filter{or, name, or, customFields, tag},
			want:        []string{"web-untagged", "db-tagged", "web-tagged", "db-custom"},
		},
		{
			description: "an operator applies to the condition after it",
			filters:     []query.Filter{name, and, tag},
			want:        []string{"web-tagged"},
		},
		{
			description: "an operator joins every condition after it until the next operator",
			filters:     []query.Filter{tag, and, name, tag},
			want:        []string{"web-tagged"},
		},
		{
			description: "and binds tighter than or",
			filters:     []query.Filter{or, name, or, customFields, and, tag},
			want:        []string{"web-untagged", "web-tagged", "db-custom"},
		},
		{
			description: "repeating the tag beside each search condition requires it",
			filters:     []query.Filter{tag, and, name, or, tag, and, customFields},
			want:        []string{"web-tagged", "db-custom"},
		},
		{
			description: "a trailing operator is ignored",
			filters:     []query.Filter{name, tag, and},
			want:        []string{"web-untagged", "db-tagged", "web-tagged", "db-custom"},
		},
		{
			description: "a group is evaluated before the operator after it",
			filters:     []query.Filter{group(name, or, customFields), and, tag},
			want:        []string{"web-tagged", "db-custom"},
		},
		{
			description: "a group after a condition joins with the operator before it",
			filters:     []query.Filter{tag, and, group(name, or, customFields)},
			want:        []string{"web-tagged", "db-custom"},
		},
		{
			description: "a group joins with or",
			filters:     []query.Filter{group(name, and, tag), or, customFields},
			want:        []string{"web-tagged", "db-custom"},
		},
		{
			description: "a group starts with or whatever the operator outside it",
			filters:     []query.Filter{tag, and, group(name, customFields)},
			want:        []string{"web-tagged", "db-custom"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			names, err := list(tc.filters)
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.want, names)
		})
	}

	invalidOnline := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "online", Operator: "bool", Value: "maybe"}}
	unknownOperator := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "name", Operator: "between", Value: "web"}}

	invalid := []struct {
		description string
		filters     []query.Filter
	}{
		{
			description: "an invalid property stays an error after a property the store skips",
			filters:     []query.Filter{invalidOnline, or, unknownOperator},
		},
		{
			description: "an invalid property inside a group is an error",
			filters:     []query.Filter{group(name, or, invalidOnline)},
		},
		{
			description: "a node of an unknown type is an error",
			filters:     []query.Filter{name, {Type: "unknown", Params: &query.FilterOperator{Name: "or"}}},
		},
		{
			description: "a property node with params of another shape is an error",
			filters:     []query.Filter{{Type: query.FilterTypeProperty, Params: &query.FilterOperator{Name: "or"}}},
		},
	}

	for _, tc := range invalid {
		t.Run(tc.description, func(t *testing.T) {
			_, err := list(tc.filters)
			assert.Error(t, err)
		})
	}

	invalidTags := query.Filter{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "tags.name", Operator: "contains", Value: float64(42)}}

	_, onlineErr := list([]query.Filter{invalidOnline})
	_, tagsErr := list([]query.Filter{invalidTags})
	require.Error(t, onlineErr)
	require.Error(t, tagsErr)
	require.NotEqual(t, onlineErr, tagsErr)

	_, err := list([]query.Filter{invalidOnline, or, invalidTags})
	assert.Equal(t, onlineErr, err, "two invalid properties return the first one's error")
}
