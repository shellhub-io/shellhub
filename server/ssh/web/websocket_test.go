package web

import (
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetToken(t *testing.T) {
	type Expected struct {
		token string
		err   error
	}

	tests := []struct {
		description string
		uri         string
		expected    Expected
	}{
		{
			description: "fail when token is not set",
			uri:         "http://localhost",
			expected: Expected{
				token: "",
				err:   ErrGetToken,
			},
		},
		{
			description: "fail when token is empty on query",
			uri:         "http://localhost?token=",
			expected: Expected{
				token: "",
				err:   ErrGetToken,
			},
		},
		{
			description: "success to get the token from query",
			uri:         "http://localhost?token=foo",
			expected: Expected{
				token: "foo",
				err:   nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(t.Context(), "", test.uri, nil)

			token, err := getToken(req)

			assert.Equal(t, test.expected.token, token)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestGetDimensions(t *testing.T) {
	type Expected struct {
		cols uint16
		rows uint16
		err  error
	}

	tests := []struct {
		description string
		uri         string
		expected    Expected
	}{
		{
			description: "fail when cols and rows is not set",
			uri:         "http://localhost",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when cols is set but rows do not",
			uri:         "http://localhost?cols=100",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when rows is set but cols do not",
			uri:         "http://localhost?rows=100",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when cols and rows are set, but empty",
			uri:         "http://localhost?cols=&rows=",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when cols is a negative value",
			uri:         "http://localhost?cols=-100&rows=100",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when rows is a negative value",
			uri:         "http://localhost?cols=100&rows=-100",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "fail when cols or rows exceed the uint16 limit",
			uri:         "http://localhost?cols=65536&rows=65536",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "success to get the cols and rows uint16 limit",
			uri:         "http://localhost?cols=65535&rows=65535",
			expected: Expected{
				cols: math.MaxUint16,
				rows: math.MaxUint16,
				err:  nil,
			},
		},
		{
			description: "success to get the cols and rows",
			uri:         "http://localhost?cols=100&rows=100",
			expected: Expected{
				cols: 100,
				rows: 100,
				err:  nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			req, _ := http.NewRequestWithContext(t.Context(), "", test.uri, nil)

			cols, rows, err := getDimensions(req)

			assert.Equal(t, test.expected.cols, cols)
			assert.Equal(t, test.expected.rows, rows)

			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestGetIP(t *testing.T) {
	forged := http.Header{"X-Real-Ip": {"198.51.100.1"}}

	t.Run("takes the address the router resolved, not the header", func(t *testing.T) {
		req := httptest.NewRequestWithContext(origin.NewContext(t.Context(), origin.Origin{IP: "203.0.113.9"}), http.MethodGet, "/", nil)
		req.Header = forged

		ip, err := getIP(req)
		require.NoError(t, err)
		assert.Equal(t, "203.0.113.9", ip)
	})

	t.Run("fails when the router resolved no origin", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.Header = forged

		ip, err := getIP(req)
		require.ErrorIs(t, err, ErrGetIP)
		assert.Empty(t, ip)
	})

	t.Run("fails when the resolved origin has no address", func(t *testing.T) {
		req := httptest.NewRequestWithContext(origin.NewContext(t.Context(), origin.Origin{}), http.MethodGet, "/", nil)

		_, err := getIP(req)
		require.ErrorIs(t, err, ErrGetIP)
	})
}
