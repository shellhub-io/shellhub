package requests_test

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstanceAPIKeyAdminID(t *testing.T) {
	cases := []struct {
		description string
		adminID     string
		valid       bool
	}{
		{"accepts the acting administrator's ID", "6f4c1b2a1e2f3a4b5c6d7e8f", true},
		{"refuses a request with no acting administrator", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			valid, _ := validator.New().Struct(&requests.CreateInstanceAPIKey{
				AdminID:   tc.adminID,
				Name:      "ci-key",
				ExpiresAt: 30,
			})
			assert.Equal(t, tc.valid, valid)
		})
	}
}
