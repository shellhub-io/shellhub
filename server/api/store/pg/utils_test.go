package pg

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromSQLError(t *testing.T) {
	tests := []struct {
		name  string
		input error

		check func(t *testing.T, result error)
	}{
		{
			name:  "nil passes through",
			input: nil,
			check: func(t *testing.T, result error) {
				t.Helper()

				assert.NoError(t, result)
			},
		},
		{
			name:  "sql.ErrNoRows maps to store.ErrNoDocuments",
			input: sql.ErrNoRows,
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				assert.ErrorIs(t, result, store.ErrNoDocuments)
			},
		},
		{
			name:  "pgconn unique_violation (23505) maps to store.ErrDuplicate",
			input: &pgconn.PgError{Code: "23505"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				assert.ErrorIs(t, result, store.ErrDuplicate)
			},
		},
		{
			name:  "23505 with users_email_key joins DuplicateFieldError with email",
			input: &pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, store.ErrDuplicate)

				var df store.DuplicateFieldError
				require.ErrorAs(t, result, &df)
				assert.Equal(t, "email", df.Field)
			},
		},
		{
			name:  "23505 with users_username_key joins DuplicateFieldError with username",
			input: &pgconn.PgError{Code: "23505", ConstraintName: "users_username_key"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, store.ErrDuplicate)

				var df store.DuplicateFieldError
				require.ErrorAs(t, result, &df)
				assert.Equal(t, "username", df.Field)
			},
		},
		{
			name:  "23505 with unrelated constraint returns bare ErrDuplicate without DuplicateFieldError",
			input: &pgconn.PgError{Code: "23505", ConstraintName: "some_other_unique_key"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, store.ErrDuplicate)

				var df store.DuplicateFieldError
				assert.NotErrorAs(t, result, &df, "expected no DuplicateFieldError for unknown constraint")
			},
		},
		{
			name:  "restrict_violation (23001) on the instance FK maps to ErrNamespaceInstanceProtected",
			input: &pgconn.PgError{Code: "23001", ConstraintName: "systems_instance_tenant_id_fkey"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, store.ErrNamespaceInstanceProtected)
				assert.NotErrorIs(t, result, store.ErrInternal)
			},
		},
		{
			name:  "foreign_key_violation (23503) on the instance FK maps to ErrNamespaceInstanceProtected",
			input: &pgconn.PgError{Code: "23503", ConstraintName: "systems_instance_tenant_id_fkey"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				assert.ErrorIs(t, result, store.ErrNamespaceInstanceProtected)
			},
		},
		{
			name:  "restrict_violation (23001) on an unrelated constraint stays internal",
			input: &pgconn.PgError{Code: "23001", ConstraintName: "some_other_fkey"},
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, store.ErrInternal)
				assert.NotErrorIs(t, result, store.ErrNamespaceInstanceProtected)
			},
		},
		{
			name:  "generic unmapped error wraps with store.ErrInternal",
			input: errors.New("some unexpected db error"),
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				assert.ErrorIs(t, result, store.ErrInternal, "expected result to wrap store.ErrInternal")
			},
		},
		{
			name:  "context.Canceled passes through unwrapped (not ErrInternal)",
			input: context.Canceled,
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, context.Canceled, "expected result to be context.Canceled")
				assert.NotErrorIs(t, result, store.ErrInternal, "context.Canceled must NOT wrap store.ErrInternal")
			},
		},
		{
			name:  "context.DeadlineExceeded passes through unwrapped (not ErrInternal)",
			input: context.DeadlineExceeded,
			check: func(t *testing.T, result error) {
				t.Helper()

				require.Error(t, result)
				require.ErrorIs(t, result, context.DeadlineExceeded, "expected result to be context.DeadlineExceeded")
				assert.NotErrorIs(t, result, store.ErrInternal, "context.DeadlineExceeded must NOT wrap store.ErrInternal")
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
