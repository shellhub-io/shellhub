package internal

import (
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestParseFilterProperty_Comparison(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		value any
		want  any
	}{
		{name: "signed integer", value: int64(-2), want: int64(-2)},
		{name: "unsigned integer", value: uint64(2), want: uint64(2)},
		{name: "float", value: float32(1.5), want: float32(1.5)},
		{name: "timestamp", value: at, want: at},
		{name: "integer string", value: "2", want: 2},
		{name: "decimal string", value: "1.5", want: float64(1.5)},
	}
	rejected := []struct {
		name    string
		value   any
		wantErr error
	}{
		{name: "non-numeric string", value: "invalid", wantErr: strconv.ErrSyntax},
		{name: "boolean", value: true, wantErr: ErrUnsupportedNumericType},
		{name: "nil", value: nil, wantErr: ErrUnsupportedNumericType},
	}
	for _, operator := range []struct{ name, sql string }{{"gt", "? > ?"}, {"lt", "? < ?"}} {
		t.Run(operator.name, func(t *testing.T) {
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					fp := &query.FilterProperty{Name: "started_at", Operator: operator.name, Value: tc.value}
					condition, args, ok, err := ParseFilterProperty(fp, "session", time.Time{})
					require.NoError(t, err)
					assert.True(t, ok)
					assert.Equal(t, operator.sql, condition)
					assert.Equal(t, []any{bun.Ident("session.started_at"), tc.want}, args)
				})
			}
			for _, tc := range rejected {
				t.Run(tc.name, func(t *testing.T) {
					fp := &query.FilterProperty{Name: "started_at", Operator: operator.name, Value: tc.value}
					condition, args, ok, err := ParseFilterProperty(fp, "session", time.Time{})
					require.ErrorIs(t, err, tc.wantErr)
					assert.False(t, ok)
					assert.Empty(t, condition)
					assert.Nil(t, args)
				})
			}
		})
	}
}
