package requests_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstanceAPIKeyUsername(t *testing.T) {
	cases := []struct {
		description string
		username    string
		valid       bool
	}{
		{"accepts the acting administrator's username", "admin", true},
		{"refuses an empty username, which several SAML users share", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			valid, _ := validator.New().Struct(&requests.CreateInstanceAPIKey{
				Username:  tc.username,
				Name:      "ci-key",
				ExpiresAt: 30,
			})
			assert.Equal(t, tc.valid, valid)
		})
	}
}
