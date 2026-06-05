package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSQLError(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		check   func(t *testing.T, result error)
	}{
		{
			name:  "nil passes through",
			input: nil,
			check: func(t *testing.T, result error) {
				assert.NoError(t, result)
			},
		},
		{
			name:  "sql.ErrNoRows maps to store.ErrNoDocuments",
			input: sql.ErrNoRows,
			check: func(t *testing.T, result error) {
				require.Error(t, result)
				assert.True(t, errors.Is(result, store.ErrNoDocuments))
			},
		},
		{
			name:  "pgconn unique_violation (23505) maps to store.ErrDuplicate",
			input: &pgconn.PgError{Code: "23505"},
			check: func(t *testing.T, result error) {
				require.Error(t, result)
				assert.True(t, errors.Is(result, store.ErrDuplicate))
			},
		},
		{
			name:  "generic unmapped error wraps with store.ErrInternal",
			input: fmt.Errorf("some unexpected db error"),
			check: func(t *testing.T, result error) {
				require.Error(t, result)
				assert.True(t, errors.Is(result, store.ErrInternal), "expected result to wrap store.ErrInternal")
			},
		},
		{
			name:  "context.Canceled passes through unwrapped (not ErrInternal)",
			input: context.Canceled,
			check: func(t *testing.T, result error) {
				require.Error(t, result)
				assert.True(t, errors.Is(result, context.Canceled), "expected result to be context.Canceled")
				assert.False(t, errors.Is(result, store.ErrInternal), "context.Canceled must NOT wrap store.ErrInternal")
			},
		},
		{
			name:  "context.DeadlineExceeded passes through unwrapped (not ErrInternal)",
			input: context.DeadlineExceeded,
			check: func(t *testing.T, result error) {
				require.Error(t, result)
				assert.True(t, errors.Is(result, context.DeadlineExceeded), "expected result to be context.DeadlineExceeded")
				assert.False(t, errors.Is(result, store.ErrInternal), "context.DeadlineExceeded must NOT wrap store.ErrInternal")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromSQLError(tt.input)
			tt.check(t, result)
		})
	}
}
