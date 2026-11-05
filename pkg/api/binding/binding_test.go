package binding

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedSources(t *testing.T) {
	cases := []struct {
		description string
		tag         reflect.StructTag
		want        []Source
	}{
		{description: "a body field names none", tag: `json:"name"`, want: nil},
		{description: "an empty value names none", tag: `param:"" json:"-"`, want: nil},
		{description: "each named source in the order of Sources", tag: `header:"X-ID" param:"id" json:"-"`, want: []Source{{Tag: "param", Name: "id"}, {Tag: "header", Name: "X-ID"}}},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.want, NamedSources(tc.tag))
		})
	}
}

func TestSourceOf(t *testing.T) {
	cases := []struct {
		description string
		tag         reflect.StructTag
		want        Source
		wantErr     string
	}{
		{description: "a json key is the body", tag: `json:"name,omitempty"`, want: Source{Tag: "json", Name: "name"}},
		{description: "json:\"-\" with a param reads the path", tag: `param:"tenant" json:"-"`, want: Source{Tag: "param", Name: "tenant"}},
		{description: "json:\"-\" with a header reads the header", tag: `header:"X-ID" json:"-"`, want: Source{Tag: "header", Name: "X-ID"}},
		{description: "json:\"-\" alone is filled by the handler", tag: `json:"-"`, want: Source{}},
		{description: "no json tag is refused", tag: `param:"tenant"`, wantErr: "no json tag"},
		{description: "an untagged field is refused", tag: ``, wantErr: "no json tag"},
		{description: "a json key with another source is refused", tag: `json:"tenant" param:"tenant"`, wantErr: "both the body and param"},
		{description: "two non-body sources are refused", tag: `param:"tenant" header:"X-Tenant-ID" json:"-"`, wantErr: "both param and header"},
		{description: "a json tag without a key is refused", tag: `json:",omitempty"`, wantErr: "names no key"},
		{description: "json:\"-\" with a known claim reads the identity", tag: `identity:"tenant" json:"-"`, want: Source{Tag: "identity", Name: "tenant"}},
		{description: "an identity tag naming no claim is refused", tag: `identity:"tenat" json:"-"`, wantErr: "names no claim"},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			got, err := SourceOf(tc.tag)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

type embeddedTenant struct {
	Tenant string `param:"tenant" json:"-"`
}

type embeddedUntagged struct {
	Tenant string `param:"tenant"`
}

type wellDeclared struct {
	embeddedTenant
	UserID   string `header:"X-ID" json:"-"`
	Name     string `json:"name"`
	Settings struct {
		Record *bool `json:"record"`
	} `json:"settings"`
}

func TestSchemaOf(t *testing.T) {
	cases := []struct {
		description string
		typ         reflect.Type
		wantErr     string
	}{
		{description: "a well declared request", typ: reflect.TypeFor[*wellDeclared]()},
		{description: "a type that is not a struct binds nothing", typ: reflect.TypeFor[map[string]any]()},
		{
			description: "an embedded struct's fields follow the rule",
			typ: reflect.TypeFor[struct {
				embeddedUntagged
			}](),
			wantErr: "embeddedUntagged.Tenant: no json tag",
		},
		{
			description: "a source tag inside a body field is refused",
			typ: reflect.TypeFor[struct {
				Nested struct {
					Tenant string `param:"tenant"`
				} `json:"nested"`
			}](),
			wantErr: "Nested.Tenant: param tag inside a body field",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := SchemaOf(tc.typ)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestSchemaForeign(t *testing.T) {
	type request struct {
		Tenant  string `param:"tenant" json:"-"`
		Name    string `json:"name"`
		Alias   string `json:"tenantid"`
		Tenant2 string `header:"X-Tenant-ID" json:"-"`
	}

	schema, err := SchemaOf(reflect.TypeFor[request]())
	require.NoError(t, err)

	var object map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(`{"TENANT":1,"name":2,"Tenant2":3,"TenantID":4,"other":5}`), &object))

	assert.Equal(t, []string{"TENANT", "Tenant2"}, schema.Foreign(object))
}

func TestMustCheck(t *testing.T) {
	assert.NotPanics(t, MustCheck[wellDeclared])
	assert.Panics(t, MustCheck[struct{ Name string }])
}
