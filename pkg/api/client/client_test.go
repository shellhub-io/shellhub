package client

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
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
			description:   "success to create a new client",
			address:       "http://localhost",
			opts:          []Opt{},
			requiredMocks: func(client *http.Client) {},
			err:           nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.cloud.shellhub.io/")
			assert.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			httpmock.ActivateNonDefault(client.http.GetClient())
			defer httpmock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

			_, err = NewClient(test.address, test.opts...)
			assert.ErrorIs(t, err, test.err)
		})
	}
}
