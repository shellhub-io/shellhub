package web

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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
			req, _ := http.NewRequest("", test.uri, nil)

			token, err := GetToken(req)

			assert.Equal(t, test.expected.token, token)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestGetDimensions(t *testing.T) {
	type Expected struct {
		cols int
		rows int
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
			description: "fail when cols or rows exceed the uint8 limit",
			uri:         "http://localhost?cols=256&rows=256",
			expected: Expected{
				cols: 0,
				rows: 0,
				err:  ErrGetDimensions,
			},
		},
		{
			description: "success to get the cols and rows unint8 limit",
			uri:         "http://localhost?cols=255&rows=255",
			expected: Expected{
				cols: 255,
				rows: 255,
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
			req, _ := http.NewRequest("", test.uri, nil)

			cols, rows, err := GetDimensions(req)

			assert.Equal(t, test.expected.cols, cols)
			assert.Equal(t, test.expected.rows, rows)

			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}

func TestGetIP(t *testing.T) {
	type Expected struct {
		ip  string
		err error
	}

	tests := []struct {
		description string
		req         *http.Request
		expected    Expected
	}{
		{
			description: "fail when IP is not set",
			req: &http.Request{
				Header: http.Header{},
			},
			expected: Expected{
				ip:  "",
				err: ErrGetIP,
			},
		},
		{
			description: "fail when IP is empty on query",
			req: &http.Request{
				Header: map[string][]string{
					"X-Real-Ip": {},
				},
			},
			expected: Expected{
				ip:  "",
				err: ErrGetIP,
			},
		},
		{
			description: "success to get the IP from query",
			req: &http.Request{
				Header: map[string][]string{
					"X-Real-Ip": {"192.168.1.1"},
				},
			},
			expected: Expected{
				ip:  "192.168.1.1",
				err: nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ip, err := GetIP(test.req)

			assert.Equal(t, test.expected.ip, ip)
			assert.ErrorIs(t, err, test.expected.err)
		})
	}
}
