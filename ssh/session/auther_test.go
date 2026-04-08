package session

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPasswordAuthEvaluate(t *testing.T) {
	cases := []struct {
		name            string
		disablePassword bool
		expectedError   error
	}{
		{
			name:            "password auth enabled",
			disablePassword: false,
			expectedError:   nil,
		},
		{
			name:            "password auth disabled",
			disablePassword: true,
			expectedError:   ErrPasswordDisabled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := &Session{
				Data: Data{
					Namespace: &models.Namespace{
						Settings: &models.NamespaceSettings{
							DisablePassword: tc.disablePassword,
						},
					},
				},
			}

			auth := AuthPassword("password")
			err := auth.Evaluate(sess)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestPublicKeyAuthEvaluate(t *testing.T) {
	cases := []struct {
		name             string
		disablePublicKey bool
		expectedError    error
	}{
		{
			name:             "public key auth disabled",
			disablePublicKey: true,
			expectedError:    ErrPublicKeyDisabled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sess := &Session{
				Data: Data{
					Device: &models.Device{
						Info: &models.DeviceInfo{
							Version: "latest",
						},
					},
					Namespace: &models.Namespace{
						Settings: &models.NamespaceSettings{
							DisablePublicKey: tc.disablePublicKey,
						},
					},
				},
			}

			auth := AuthPublicKey(nil)
			err := auth.Evaluate(sess)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}
