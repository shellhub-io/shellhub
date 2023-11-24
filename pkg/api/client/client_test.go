package client

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	ErrFromOption := errors.New("error from option")

	tests := []struct {
		description   string
		address       string
		opts          []Opt
		requiredMocks func(client *http.Client)
		err           error
	}{
		{
			description:   "failed to create when address is invalid",
			address:       "localhost",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           ErrParseAddress,
		},
		{
			description: "failed to create when some option return error",
			address:     "http://localhost",
			opts: []Opt{
				func(c *client) error {
					return ErrFromOption
				},
			},
			requiredMocks: func(client *http.Client) {},
			err:           ErrFromOption,
		},
		{
			description:   "success to create a new client with 127.0.0.1 in http",
			address:       "http://127.0.0.1",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with 127.0.0.1 in https",
			address:       "https://127.0.0.1",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with 127.0.0.1 in http with port",
			address:       "http://127.0.0.1:80",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with 127.0.0.1 in https with port",
			address:       "https://127.0.0.1:443",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with localhost in http",
			address:       "http://localhost",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with localhost in https",
			address:       "https://localhost",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with localhost in http with port",
			address:       "http://localhost:80",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with localhost in https with port",
			address:       "https://localhost:443",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with cloud.shellhub.io in https",
			address:       "https://cloud.shellhub.io",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
		{
			description:   "success to create a new client with cloud.shellhub.io in https with port",
			address:       "https://cloud.shellhub.io:443",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			_, err := NewClient(test.address, test.opts...)
			assert.ErrorIs(t, err, test.err)
		})
	}
}
