package internal

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestMapFieldToColumn(t *testing.T) {
	cases := []struct {
		description string
		field       string
		want        string
	}{
		{
			// device_uid is NOT a global alias — the mapping to device_id is
			// session-specific and applied by ParseFilterProperty when tableAlias
			// is "session".  mapFieldToColumn must not alter it globally.
			description: "device_uid passes through unchanged (not a global alias)",
			field:       "device_uid",
			want:        "device_uid",
		},
		{
			description: "legacy mongo field info.platform maps to platform",
			field:       "info.platform",
			want:        "platform",
		},
		{
			description: "unknown field passes through unchanged",
			field:       "name",
			want:        "name",
		},
		{
			description: "identity.mac maps to mac",
			field:       "identity.mac",
			want:        "mac",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			got := mapFieldToColumn(tc.field)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestParseFilterProperty_DeviceUID verifies that the device_uid→device_id
// column alias is applied only in the session context (tableAlias == "session"),
// not globally.  If applied globally, a future endpoint that exposes device_uid
// for a table with no device_id column would silently produce wrong SQL instead
// of a column-not-found error.
func TestParseFilterProperty_DeviceUID(t *testing.T) {
	cases := []struct {
		description string
		tableAlias  string
		wantCol     bun.Ident // the column identifier expected in args[0]
	}{
		{
			description: "session context: device_uid maps to session.device_id",
			tableAlias:  "session",
			wantCol:     bun.Ident("session.device_id"),
		},
		{
			// In non-session contexts device_uid is an unknown alias and must
			// NOT be silently remapped — it should produce the literal column
			// name so that the database rejects it with a clear error if the
			// column does not exist.
			description: "non-session context: device_uid passes through unchanged",
			tableAlias:  "device",
			wantCol:     bun.Ident("device.device_uid"),
		},
		{
			description: "empty alias: device_uid passes through unchanged",
			tableAlias:  "",
			wantCol:     bun.Ident("device_uid"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			fp := &query.FilterProperty{Name: "device_uid", Operator: "eq", Value: "abc"}
			sqlCond, args, ok, err := ParseFilterProperty(fp, tc.tableAlias)
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, "? = ?", sqlCond)
			require.Len(t, args, 2)
			assert.Equal(t, tc.wantCol, args[0])
			assert.Equal(t, "abc", args[1])
		})
	}
}

func TestFromActiveFilter(t *testing.T) {
	const deviceExistsSQL = `EXISTS (SELECT 1 FROM "active_sessions" JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id" WHERE "sessions"."device_id" = "device"."id")`
	const deviceNotExistsSQL = `NOT EXISTS (SELECT 1 FROM "active_sessions" JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id" WHERE "sessions"."device_id" = "device"."id")`
	const sessionExistsSQL = `EXISTS (SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`
	const sessionNotExistsSQL = `NOT EXISTS (SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`

	cases := []struct {
		description string
		value       any
		tableAlias  string
		wantSQL     string
		wantArgs    []any
		wantOk      bool
		wantErr     error
	}{
		{
			description: "device context: bool true produces EXISTS device subquery",
			value:       true,
			tableAlias:  "",
			wantSQL:     deviceExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "device context: bool false produces NOT EXISTS device subquery",
			value:       false,
			tableAlias:  "",
			wantSQL:     deviceNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "device context: string true produces EXISTS device subquery",
			value:       "true",
			tableAlias:  "",
			wantSQL:     deviceExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "device context: string false produces NOT EXISTS device subquery",
			value:       "false",
			tableAlias:  "",
			wantSQL:     deviceNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "session context: bool true produces EXISTS session subquery",
			value:       true,
			tableAlias:  "session",
			wantSQL:     sessionExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "session context: bool false produces NOT EXISTS session subquery",
			value:       false,
			tableAlias:  "session",
			wantSQL:     sessionNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "session context: float64 1 (JSON number) produces EXISTS session subquery",
			value:       float64(1),
			tableAlias:  "session",
			wantSQL:     sessionExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "session context: float64 0 (JSON number) produces NOT EXISTS session subquery",
			value:       float64(0),
			tableAlias:  "session",
			wantSQL:     sessionNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "device context: float64 1 (JSON number) produces EXISTS device subquery",
			value:       float64(1),
			tableAlias:  "",
			wantSQL:     deviceExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			// int is not produced by JSON decoding (JSON numbers decode to float64),
			// so int remains an unsupported type.
			description: "unsupported type int returns error",
			value:       42,
			tableAlias:  "",
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     ErrUnsupportedBoolType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sql, args, ok, err := fromActiveFilter(tc.value, tc.tableAlias)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantSQL, sql)
			assert.Equal(t, tc.wantArgs, args)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

// TestParseFilterProperty_Bool verifies that fromBool correctly handles all
// value types accepted from client JSON, including the float64 case (JSON
// numbers always decode to float64, never int).
func TestParseFilterProperty_Bool(t *testing.T) {
	cases := []struct {
		description string
		value       any
		wantBool    bool
		wantOk      bool
		wantErr     bool
	}{
		{description: "bool true", value: true, wantBool: true, wantOk: true},
		{description: "bool false", value: false, wantBool: false, wantOk: true},
		{description: "string 1", value: "1", wantBool: true, wantOk: true},
		{description: "string false", value: "false", wantBool: false, wantOk: true},
		{description: "float64 nonzero (JSON number)", value: float64(1), wantBool: true, wantOk: true},
		{description: "float64 zero (JSON number)", value: float64(0), wantBool: false, wantOk: true},
		// int is not produced by JSON decoding but is still handled by fromBool
		// for programmatic callers.
		{description: "int nonzero (programmatic, not JSON)", value: 1, wantBool: true, wantOk: true},
		{description: "invalid string", value: "yes", wantOk: false, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			fp := &query.FilterProperty{Name: "closed", Operator: "bool", Value: tc.value}
			sqlCond, args, ok, err := ParseFilterProperty(fp, "session")
			assert.Equal(t, tc.wantOk, ok)

			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, "? = ?", sqlCond)
			require.Len(t, args, 2)
			// args[0] is the qualified column identifier; args[1] is the resolved bool.
			assert.Equal(t, bun.Ident("session.closed"), args[0])
			assert.Equal(t, tc.wantBool, args[1])
		})
	}
}

func TestFromOnlineFilter(t *testing.T) {
	const onlineSQL = `("device"."disconnected_at" IS NULL AND "device"."last_seen" > ?)`
	const offlineSQL = `("device"."disconnected_at" IS NOT NULL OR "device"."last_seen" <= ?)`

	cases := []struct {
		description string
		value       any
		wantSQL     string
		wantOk      bool
		wantErr     error
	}{
		{
			description: "bool true produces online SQL",
			value:       true,
			wantSQL:     onlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "bool false produces offline SQL",
			value:       false,
			wantSQL:     offlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "string true produces online SQL",
			value:       "true",
			wantSQL:     onlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "string false produces offline SQL",
			value:       "false",
			wantSQL:     offlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			// JSON numbers always decode to float64; nonzero means online.
			description: "float64 nonzero (JSON number) produces online SQL",
			value:       float64(1),
			wantSQL:     onlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			description: "float64 zero (JSON number) produces offline SQL",
			value:       float64(0),
			wantSQL:     offlineSQL,
			wantOk:      true,
			wantErr:     nil,
		},
		{
			// int is not produced by JSON decoding (JSON numbers decode to float64),
			// so int remains an unsupported type.
			description: "unsupported type int returns ErrUnsupportedBoolType",
			value:       42,
			wantSQL:     "",
			wantOk:      false,
			wantErr:     ErrUnsupportedBoolType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sql, args, ok, err := fromOnlineFilter(tc.value)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantErr, err)

			if tc.wantOk {
				assert.Equal(t, tc.wantSQL, sql)
				require.Len(t, args, 1)
			}
		})
	}
}

func TestParseFilterProperty_Active(t *testing.T) {
	const deviceExistsSQL = `EXISTS (SELECT 1 FROM "active_sessions" JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id" WHERE "sessions"."device_id" = "device"."id")`
	const deviceNotExistsSQL = `NOT EXISTS (SELECT 1 FROM "active_sessions" JOIN "sessions" ON "sessions"."id" = "active_sessions"."session_id" WHERE "sessions"."device_id" = "device"."id")`
	const sessionExistsSQL = `EXISTS (SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`
	const sessionNotExistsSQL = `NOT EXISTS (SELECT 1 FROM "active_sessions" WHERE "active_sessions"."session_id" = "session"."id")`

	cases := []struct {
		description string
		fp          *query.FilterProperty
		tableAlias  string
		wantSQL     string
		wantArgs    []any
		wantOk      bool
		wantErr     bool
	}{
		{
			description: "device context: active=true routes to EXISTS device subquery",
			fp:          &query.FilterProperty{Name: "active", Value: true},
			tableAlias:  "device",
			wantSQL:     deviceExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "device context: active=false routes to NOT EXISTS device subquery",
			fp:          &query.FilterProperty{Name: "active", Value: false},
			tableAlias:  "device",
			wantSQL:     deviceNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "session context: active=true routes to EXISTS session subquery",
			fp:          &query.FilterProperty{Name: "active", Value: true},
			tableAlias:  "session",
			wantSQL:     sessionExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "session context: active=false routes to NOT EXISTS session subquery",
			fp:          &query.FilterProperty{Name: "active", Value: false},
			tableAlias:  "session",
			wantSQL:     sessionNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "session context: float64 1 (JSON number) produces EXISTS session subquery",
			fp:          &query.FilterProperty{Name: "active", Value: float64(1)},
			tableAlias:  "session",
			wantSQL:     sessionExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			description: "session context: float64 0 (JSON number) produces NOT EXISTS session subquery",
			fp:          &query.FilterProperty{Name: "active", Value: float64(0)},
			tableAlias:  "session",
			wantSQL:     sessionNotExistsSQL,
			wantArgs:    nil,
			wantOk:      true,
			wantErr:     false,
		},
		{
			// int is not produced by JSON decoding (JSON numbers decode to float64),
			// so int remains unsupported.
			description: "active with unsupported value type int returns error",
			fp:          &query.FilterProperty{Name: "active", Value: 99},
			tableAlias:  "device",
			wantSQL:     "",
			wantArgs:    nil,
			wantOk:      false,
			wantErr:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			sql, args, ok, err := ParseFilterProperty(tc.fp, tc.tableAlias)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.wantSQL, sql)
			assert.Equal(t, tc.wantArgs, args)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
