//go:build freebsd

package osauth

import (
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/stretchr/testify/assert"
)

const testHash = "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN."

func setNow(t *testing.T, seconds int64) {
	t.Helper()

	clockMock := clockmock.NewMockClock(t)
	clockMock.On("Now").Return(time.Unix(seconds, 0)).Maybe()

	previous := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	t.Cleanup(func() { clock.DefaultBackend = previous })
}

func masterPasswdLine(username, change, expire string) string {
	return username + ":" + testHash + ":1001:1001::" + change + ":" + expire + ":User:/home/user:/bin/sh\n"
}

func TestAuthUserFromShadowExpiry(t *testing.T) {
	setNow(t, 1800000000)

	tests := []struct {
		name   string
		change string
		expire string
		want   bool
	}{
		{
			name:   "no change or expire time",
			change: "0",
			expire: "0",
			want:   true,
		},
		{
			name:   "change and expire in the future",
			change: "1800000001",
			expire: "1800000001",
			want:   true,
		},
		{
			name:   "account expired",
			change: "0",
			expire: "1800000000",
			want:   false,
		},
		{
			name:   "password past its change time",
			change: "1800000000",
			expire: "0",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masterPasswd := strings.NewReader(masterPasswdLine("user", tt.change, tt.expire))

			assert.Equal(t, tt.want, AuthUserFromShadow("user", "123", masterPasswd))
		})
	}
}

func TestAuthUserFromShadowSkipsMalformedLine(t *testing.T) {
	setNow(t, 1800000000)

	masterPasswd := masterPasswdLine("bad", "0", "never") + masterPasswdLine("good", "0", "0")

	assert.False(t, AuthUserFromShadow("bad", "123", strings.NewReader(masterPasswd)))
	assert.True(t, AuthUserFromShadow("good", "123", strings.NewReader(masterPasswd)))
}

func TestAccountExpiredFromShadow(t *testing.T) {
	setNow(t, 1800000000)

	masterPasswd := masterPasswdLine("expired", "0", "1800000000") +
		masterPasswdLine("changed", "1700000000", "0") +
		masterPasswdLine("active", "0", "1800000001")

	assert.True(t, AccountExpiredFromShadow("expired", strings.NewReader(masterPasswd)))
	assert.False(t, AccountExpiredFromShadow("changed", strings.NewReader(masterPasswd)))
	assert.False(t, AccountExpiredFromShadow("active", strings.NewReader(masterPasswd)))
	assert.False(t, AccountExpiredFromShadow("missing", strings.NewReader(masterPasswd)))
}

func TestVerifyPasswordHashEmptyHash(t *testing.T) {
	t.Setenv("SHELLHUB_PERMIT_EMPTY_PASSWORDS", "")
	assert.False(t, VerifyPasswordHash("", ""))

	t.Setenv("SHELLHUB_PERMIT_EMPTY_PASSWORDS", "true")
	assert.True(t, VerifyPasswordHash("", ""))
}
