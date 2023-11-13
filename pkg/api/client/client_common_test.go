package client

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/shellhub-io/shellhub/pkg/models"
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
			cli, err := NewClient("https://www.shellhub.io/")
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

func TestListDevices(t *testing.T) {
	type Expected struct {
		devices []models.Device
		err     error
	}

	tests := []struct {
		description   string
		requiredMocks func(client *http.Client)
		expected      Expected
	}{
		{
			description: "success to list devices when its list is empty",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(200, []models.Device{{}})
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: []models.Device{{}},
				err:     nil,
			},
		},
		{
			description: "success to list devices when its list is not empty",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(200, []models.Device{
					{
						UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					},
					{
						UID: "6fdc847ff6701685268393b0e418e75c95cf94aa4415bcdcd65cd8e2b7345b0d",
					},
				})
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: []models.Device{
					{
						UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					},
					{
						UID: "6fdc847ff6701685268393b0e418e75c95cf94aa4415bcdcd65cd8e2b7345b0d",
					},
				},
				err: nil,
			},
		},
		{
			description: "success to list devices after retry",
			requiredMocks: func(client *http.Client) {
				fail := httpmock.NewErrorResponder(errors.New("error on request"))
				success, _ := httpmock.NewJsonResponder(200, []models.Device{
					{
						UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					},
					{
						UID: "6fdc847ff6701685268393b0e418e75c95cf94aa4415bcdcd65cd8e2b7345b0d",
					},
				})

				responder := fail.
					Then(fail).
					Then(fail).
					Then(success)
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: []models.Device{
					{
						UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
					},
					{
						UID: "6fdc847ff6701685268393b0e418e75c95cf94aa4415bcdcd65cd8e2b7345b0d",
					},
				},
				err: nil,
			},
		},
		{
			description: "failed when resource is not found",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(404, nil)
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: nil,
				err:     ErrNotFound,
			},
		},
		{
			description: "failed when request is missformated",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(400, nil)
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: nil,
				err:     ErrBadRequest,
			},
		},
		{
			description: "failed when request return an unmaped status code",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(418, nil)
				httpmock.RegisterResponder("GET", "/api/devices", responder)
			},
			expected: Expected{
				devices: nil,
				err:     errors.Join(ErrUnknown, errors.New("418")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.shellhub.io/")
			assert.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			httpmock.ActivateNonDefault(client.http.GetClient())
			defer httpmock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

			list, err := cli.ListDevices()
			assert.Equal(t, test.expected, Expected{list, err})
		})
	}
}

func TestGetDevice(t *testing.T) {
	type Expected struct {
		device *models.Device
		err    error
	}

	tests := []struct {
		description   string
		uid           string
		requiredMocks func(client *http.Client)
		expected      Expected
	}{
		{
			description: "success to get device",
			uid:         "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(200, models.Device{
					UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
				})
				httpmock.RegisterResponder("GET", fmt.Sprintf("/api/devices/%s", "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117"), responder)
			},
			expected: Expected{
				device: &models.Device{
					UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
				},
				err: nil,
			},
		},
		{
			description: "success to get device after retry",
			uid:         "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
			requiredMocks: func(client *http.Client) {
				fail := httpmock.NewErrorResponder(errors.New("error on request"))
				success, _ := httpmock.NewJsonResponder(200, models.Device{
					UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
				})

				responder := fail.
					Then(fail).
					Then(fail).
					Then(success)
				httpmock.RegisterResponder("GET", fmt.Sprintf("/api/devices/%s", "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117"), responder)
			},
			expected: Expected{
				device: &models.Device{
					UID: "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
				},
				err: nil,
			},
		},
		{
			description: "failed when device is not found",
			uid:         "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(404, nil)
				httpmock.RegisterResponder("GET", fmt.Sprintf("/api/devices/%s", "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117"), responder)
			},
			expected: Expected{
				device: nil,
				err:    ErrNotFound,
			},
		},
		{
			description: "failed when device request is missformated",
			uid:         "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(400, nil)
				httpmock.RegisterResponder("GET", fmt.Sprintf("/api/devices/%s", "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117"), responder)
			},
			expected: Expected{
				device: nil,
				err:    ErrBadRequest,
			},
		},
		{
			description: "failed when device request return an unmaped error code",
			uid:         "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117",
			requiredMocks: func(client *http.Client) {
				responder, _ := httpmock.NewJsonResponder(418, nil)
				httpmock.RegisterResponder("GET", fmt.Sprintf("/api/devices/%s", "3a471bd84c88b28c4e4f8e27caee40e7b14798325e6dd85aa62d54e27fd11117"), responder)
			},
			expected: Expected{
				device: nil,
				err:    errors.Join(ErrUnknown, errors.New("418")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			cli, err := NewClient("https://www.shellhub.io/")
			assert.NoError(t, err)

			client, ok := cli.(*client)
			assert.True(t, ok)

			httpmock.ActivateNonDefault(client.http.GetClient())
			defer httpmock.DeactivateAndReset()

			test.requiredMocks(client.http.GetClient())

			device, err := cli.GetDevice(test.uid)
			assert.Equal(t, test.expected, Expected{device, err})
		})
	}
}
