package query

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorRoundTrip(t *testing.T) {
	cases := []struct {
		description string
		cursor      Cursor
	}{
		{
			description: "keeps microsecond precision",
			cursor:      Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 123456000, time.UTC), ID: "abc"},
		},
		{
			description: "keeps an id with separators in it",
			cursor:      Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "a|b,c\"d"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			encoded := tc.cursor.String()
			assert.NotContains(t, encoded, "=")
			assert.NotContains(t, encoded, "+")
			assert.NotContains(t, encoded, "/")

			decoded, err := ParseCursor(encoded)
			require.NoError(t, err)
			assert.True(t, tc.cursor.CreatedAt.Equal(decoded.CreatedAt))
			assert.Equal(t, tc.cursor.ID, decoded.ID)
		})
	}
}

func TestParseCursorRejectsWhatTheServerDidNotIssue(t *testing.T) {
	encode := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

	cases := []struct {
		description string
		cursor      string
	}{
		{description: "empty", cursor: ""},
		{description: "not base64url", cursor: "!!!"},
		{description: "standard base64 padding", cursor: base64.StdEncoding.EncodeToString([]byte(`{"c":1768046400000000,"i":"abc"}`))},
		{description: "not json", cursor: encode("not-json")},
		{description: "missing id", cursor: encode(`{"c":1768046400000000}`)},
		{description: "missing time", cursor: encode(`{"i":"abc"}`)},
		{description: "time before the epoch", cursor: encode(`{"c":-1,"i":"abc"}`)},
		{description: "time as text", cursor: encode(`{"c":"2026-01-10T12:00:00Z","i":"abc"}`)},
		{description: "unknown field", cursor: encode(`{"c":1768046400000000,"i":"abc","x":1}`)},
		{description: "trailing data", cursor: encode(`{"c":1768046400000000,"i":"abc"}{}`)},
		{description: "too long", cursor: encode(`{"c":1768046400000000,"i":"` + strings.Repeat("a", 1024) + `"}`)},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			_, err := ParseCursor(tc.cursor)
			assert.ErrorIs(t, err, ErrInvalidCursor)
		})
	}
}

func TestCursorPaginatorNormalize(t *testing.T) {
	cases := []struct {
		description string
		size        int64
		maxSize     int64
		expected    int64
	}{
		{description: "takes the default when missing", size: 0, maxSize: 1000, expected: DefaultPageSize},
		{description: "raises a negative size to the minimum", size: -5, maxSize: 1000, expected: MinPageSize},
		{description: "lowers a size above the maximum to it", size: 5000, maxSize: 1000, expected: 1000},
		{description: "keeps a size in range", size: 250, maxSize: 1000, expected: 250},
		{description: "lowers the default when the maximum is below it", size: 0, maxSize: 50, expected: 50},
		{description: "takes the default maximum when the maximum is not positive", size: 5000, maxSize: 0, expected: DefaultMaxPageSize},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			p := CursorPaginator{Size: tc.size}
			p.Normalize(tc.maxSize)
			assert.Equal(t, tc.expected, p.Size)
		})
	}
}
