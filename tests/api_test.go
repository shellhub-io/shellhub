package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func ReadToString(reader io.Reader, dst *string) error {
	buf := new(strings.Builder)

	_, err := io.Copy(buf, reader)
	if err != nil {
		return err
	}

	*dst = buf.String()

	return nil
}

func TestGettingStarted(t *testing.T) {
	t.Parallel()

	type ExpectedCommand struct {
		msg  string
		code int
	}

	type Expected struct {
		userCmd      *ExpectedCommand
		namespaceCmd *ExpectedCommand
	}

	cases := []struct {
		description string
		test        func(context.Context, *environment.Environment) (*Expected, error)
		expected    Expected
	}{
		{
			description: "succeeds",
			test: func(ctx context.Context, env *environment.Environment) (*Expected, error) {
				cli := env.GetService(environment.ServiceCLI)

				// Try to create a new user
				code, reader, err := cli.Exec(ctx, strings.Split("./cli user create john_doe secret john.doe@test.com", " "))
				if err != nil {
					return nil, err
				}

				userCmd := &ExpectedCommand{
					code: code,
					msg:  "",
				}

				if err := ReadToString(reader, &userCmd.msg); err != nil {
					return nil, err
				}

				logrus.Info(userCmd.msg)

				// Try to create a new namespace
				code, reader, err = cli.Exec(ctx, strings.Split("./cli namespace create dev john_doe 00000000-0000-4000-0000-000000000000", " "))
				if err != nil {
					return nil, err
				}

				namespaceCmd := &ExpectedCommand{
					code: code,
					msg:  "",
				}

				if err := ReadToString(reader, &namespaceCmd.msg); err != nil {
					return nil, err
				}

				logrus.Info(namespaceCmd.msg)

				auth := new(models.UserAuthResponse)
				_, err = env.Request().
					SetBody(map[string]string{
						"username": "john_doe",
						"password": "secret",
					}).
					SetResult(auth).
					Post("/api/login")
				if err != nil {
					return nil, err
				}

				devices := make([]models.Device, 1)
				assert.EventuallyWithT(
					t,
					func(collect *assert.CollectT) {
						res, err := env.Request().
							SetHeader("authorization", "Bearer "+auth.Token).
							SetResult(&devices).
							Get("/api/devices?status=pending")

						assert.Equal(collect, 200, res.StatusCode())
						assert.NoError(collect, err)
						assert.Len(collect, devices, 1)
					},
					30*time.Second,
					time.Second,
				)

				_, err = env.Request().
					SetHeader("authorization", "Bearer "+auth.Token).
					Patch(fmt.Sprintf("/api/devices/%s/accept", devices[0].UID))
				if err != nil {
					return nil, err
				}

				return &Expected{
					userCmd:      userCmd,
					namespaceCmd: namespaceCmd,
				}, nil
			},
			expected: Expected{
				userCmd: &ExpectedCommand{
					code: 0,
					msg:  "\nUsername: john_doe\nEmail: john.doe@test.com\n",
				},
				namespaceCmd: &ExpectedCommand{
					code: 0,
					msg:  "Namespace created successfully\nNamespace: dev\nTenant: 00000000-0000-4000-0000-000000000000\nOwner:", // TODO: how can we assert the Owner ID?
				},
			},
		},
	}

	builder := environment.New(t)

	for _, tt := range cases {
		tc := tt

		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			env, cleanup := builder.Clone(t).Start(ctx)
			defer cleanup()

			actual, err := tc.test(ctx, env)
			if !assert.NoError(t, err) {
				t.Fatal(err)
			}

			assert.Contains(t, actual.userCmd.msg, tc.expected.userCmd.msg)
			assert.Equal(t, actual.userCmd.code, tc.expected.userCmd.code)

			assert.Contains(t, actual.namespaceCmd.msg, tc.expected.namespaceCmd.msg)
			assert.Equal(t, actual.namespaceCmd.code, tc.expected.namespaceCmd.code)
		})
	}
}
