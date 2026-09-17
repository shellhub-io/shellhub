//go:build !freebsd

package osauth

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyPasswordHashPass(t *testing.T) {
	hashPassword := "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN." //nolint:gosec
	passwd := "123"

	result := VerifyPasswordHash(hashPassword, passwd)

	assert.True(t, result)
}

func TestVerifyPasswordHashFail(t *testing.T) {
	hashPassword := "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN." //nolint:gosec
	passwd := "test"

	result := VerifyPasswordHash(hashPassword, passwd)

	assert.False(t, result)
}

func TestVerifyPasswordHashMD5Pass(t *testing.T) {
	hashPassword := "$1$YW4a91HG$31CtH9bzW/oyJ1VOD.H/d/" //nolint:gosec
	passwd := "test"

	result := VerifyPasswordHash(hashPassword, passwd)

	assert.True(t, result)
}

func TestVerifyPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{
			name:     "sha512 correct",
			hash:     "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN.",
			password: "123",
			want:     true,
		},
		{
			name:     "sha512 incorrect",
			hash:     "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN.",
			password: "test",
			want:     false,
		},
		{
			name:     "md5 correct",
			hash:     "$1$YW4a91HG$31CtH9bzW/oyJ1VOD.H/d/",
			password: "test",
			want:     true,
		},
		{
			name:     "empty hash",
			hash:     "",
			password: "any",
			want:     false,
		},
		{
			name:     "special marker bang",
			hash:     "!",
			password: "pass",
			want:     false,
		},
		{
			name:     "special marker star",
			hash:     "*",
			password: "pass",
			want:     false,
		},
		{
			name:     "locked prefix",
			hash:     "!$6$blah",
			password: "pass",
			want:     false,
		},
		{
			name:     "unsupported algo",
			hash:     "$z$invalid$hash",
			password: "pass",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPasswordHash(tt.hash, tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVerifyPasswordHash_YescryptInvocation(t *testing.T) {
	yesHash := "$y$e0801$w1Jl9GJH1j4h0w==$Wj2b7m2vWw2m3l1iQe8qvQ=="
	_ = VerifyPasswordHash(yesHash, "password")
}

//nolint:gosec
const passwd = `root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
bin:x:2:2:bin:/bin:/usr/sbin/nologin
sys:x:3:3:sys:/dev:/usr/sbin/nologin
sync:x:4:65534:sync:/bin:/bin/sync
nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin
user1:x:1000:1000:User One:/home/user1:/bin/bash
user2:x:1001:1001:User Two:/home/user2:/bin/sh`

func TestPasswdReader(t *testing.T) {
	reader := strings.NewReader(passwd)

	users, err := parsePasswdReader(reader)
	require.NoError(t, err)
	assert.Len(t, users, 8)

	tests := []struct {
		name     string
		username string
		want     struct {
			Password string
			UID      uint32
			GID      uint32
			Shell    string
		}
	}{
		{
			name:     "root user",
			username: "root",
			want: struct {
				Password string
				UID      uint32
				GID      uint32
				Shell    string
			}{
				Password: "x",
				UID:      0,
				GID:      0,
				Shell:    "/bin/bash",
			},
		},
		{
			name:     "user1",
			username: "user1",
			want: struct {
				Password string
				UID      uint32
				GID      uint32
				Shell    string
			}{
				Password: "x",
				UID:      1000,
				GID:      1000,
				Shell:    "/bin/bash",
			},
		},
		{
			name:     "user2",
			username: "user2",
			want: struct {
				Password string
				UID      uint32
				GID      uint32
				Shell    string
			}{
				Password: "x",
				UID:      1001,
				GID:      1001,
				Shell:    "/bin/sh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, ok := users[tt.username]
			assert.True(t, ok)
			assert.Equal(t, tt.want.Password, user.Password)
			assert.Equal(t, tt.want.UID, user.UID)
			assert.Equal(t, tt.want.GID, user.GID)
			assert.Equal(t, tt.want.Shell, user.Shell)
		})
	}
}

const testHash = "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN."

func setToday(t *testing.T, days int64) {
	t.Helper()

	clockMock := clockmock.NewMockClock(t)
	clockMock.On("Now").Return(time.Unix(days*secondsPerDay+12*60*60, 0).UTC()).Maybe()

	previous := clock.DefaultBackend
	clock.DefaultBackend = clockMock
	t.Cleanup(func() { clock.DefaultBackend = previous })
}

func TestAuthUserFromShadowAging(t *testing.T) {
	setToday(t, 20000)

	tests := []struct {
		name     string
		aging    string
		password string
		want     bool
	}{
		{
			name:     "aging disabled",
			aging:    "::::::",
			password: "123",
			want:     true,
		},
		{
			name:     "aging fields blank with spaces",
			aging:    " : : : : : : ",
			password: "123",
			want:     true,
		},
		{
			name:     "within password maximum",
			aging:    "19990:0:99999:7:::",
			password: "123",
			want:     true,
		},
		{
			name:     "account expires tomorrow",
			aging:    "19990:0:99999:7::20001:",
			password: "123",
			want:     true,
		},
		{
			name:     "account expires today",
			aging:    "19990:0:99999:7::20000:",
			password: "123",
			want:     false,
		},
		{
			name:     "account expired",
			aging:    "19990:0:99999:7::19999:",
			password: "123",
			want:     false,
		},
		{
			name:     "account expired on epoch day zero",
			aging:    "19990:0:99999:7::0:",
			password: "123",
			want:     false,
		},
		{
			name:     "negative expire disables expiry",
			aging:    "19990:0:99999:7::-5:",
			password: "123",
			want:     true,
		},
		{
			name:     "password one day before maximum",
			aging:    "19911:0:90:7:::",
			password: "123",
			want:     true,
		},
		{
			name:     "password reaches maximum today",
			aging:    "19910:0:90:7:::",
			password: "123",
			want:     false,
		},
		{
			name:     "password aged out and inactive",
			aging:    "19900:0:90:7:5::",
			password: "123",
			want:     false,
		},
		{
			name:     "maximum set without last change",
			aging:    ":0:90:7:::",
			password: "123",
			want:     true,
		},
		{
			name:     "password change forced",
			aging:    "0:0:99999:7:::",
			password: "123",
			want:     false,
		},
		{
			name:     "last change in the future",
			aging:    "20010:0:5:7:::",
			password: "123",
			want:     true,
		},
		{
			name:     "wrong password on valid account",
			aging:    "19990:0:99999:7:::",
			password: "wrong",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shadow := strings.NewReader("user:" + testHash + ":" + tt.aging + "\n")

			assert.Equal(t, tt.want, AuthUserFromShadow("user", tt.password, shadow))
		})
	}
}

func TestAuthUserFromShadowSkipsMalformedLine(t *testing.T) {
	setToday(t, 20000)

	shadow := "bad:" + testHash + ":19990:0:99999:7::never:\n" +
		"good:" + testHash + ":19990:0:99999:7:::\n"

	assert.False(t, AuthUserFromShadow("bad", "123", strings.NewReader(shadow)))
	assert.True(t, AccountExpiredFromShadow("bad", strings.NewReader(shadow)))
	assert.True(t, AuthUserFromShadow("good", "123", strings.NewReader(shadow)))
}

func TestAccountExpiredFromShadow(t *testing.T) {
	setToday(t, 20000)

	tests := []struct {
		name     string
		shadow   string
		username string
		want     bool
	}{
		{
			name:     "account not expired",
			shadow:   "user:*:19990:0:99999:7::20001:\n",
			username: "user",
			want:     false,
		},
		{
			name:     "account expired",
			shadow:   "user:*:19990:0:99999:7::20000:\n",
			username: "user",
			want:     true,
		},
		{
			name:     "forced password change does not expire the account",
			shadow:   "user:*:0:0:90:7:::\n",
			username: "user",
			want:     false,
		},
		{
			name:     "account without an entry",
			shadow:   "other:*:19990:0:99999:7::19000:\n",
			username: "user",
			want:     false,
		},
		{
			name:     "account line with the wrong field count",
			shadow:   "user:*\n",
			username: "user",
			want:     true,
		},
		{
			name:     "account line with a malformed expire field",
			shadow:   "user:*:19990:0:99999:7::20000x:\n",
			username: "user",
			want:     true,
		},
		{
			name:     "malformed line shadows a later valid one",
			shadow:   "user:*:19990:0:99999:7::20000x:\nuser:*:19990:0:99999:7:::\n",
			username: "user",
			want:     true,
		},
		{
			name:     "another account's line is malformed",
			shadow:   "other:*\nuser:*:19990:0:99999:7:::\n",
			username: "user",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AccountExpiredFromShadow(tt.username, strings.NewReader(tt.shadow)))
		})
	}
}

func TestParseErrorsDoNotEchoFileContent(t *testing.T) {
	const poison = "SUPERSECRET-injected-by-a-hostile-image"

	tests := []struct {
		name     string
		data     string
		parse    func(io.Reader) error
		wantLine string
	}{
		{
			name: "passwd uid field",
			data: "root:x:0:0:root:/root:/bin/bash\nbad:x:" + poison + ":0::/:/bin/sh\n",
			parse: func(r io.Reader) error {
				_, err := parsePasswdReader(r)

				return err
			},
			wantLine: "passwd line 2",
		},
		{
			name: "passwd gid field",
			data: "bad:x:0:" + poison + "::/:/bin/sh\n",
			parse: func(r io.Reader) error {
				_, err := parsePasswdReader(r)

				return err
			},
			wantLine: "passwd line 1",
		},
		{
			name: "passwd field count",
			data: "bad:x:0:0:" + poison + "\n",
			parse: func(r io.Reader) error {
				_, err := parsePasswdReader(r)

				return err
			},
			wantLine: "passwd line 1",
		},
		{
			name: "group gid field",
			data: "root:x:0:root\nbad:x:" + poison + ":member\n",
			parse: func(r io.Reader) error {
				_, err := parseGroupReader(r)

				return err
			},
			wantLine: "group line 2",
		},
		{
			name: "shadow field count",
			data: "bad:" + poison + "\n",
			parse: func(r io.Reader) error {
				line, err := io.ReadAll(r)
				if err != nil {
					return err
				}

				_, err = parseShadowLine(string(line))

				return err
			},
			wantLine: "wrong number of fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parse(strings.NewReader(tt.data))

			require.Error(t, err)
			assert.NotContains(t, err.Error(), poison)
			assert.Contains(t, err.Error(), tt.wantLine)
		})
	}
}
