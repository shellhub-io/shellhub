//go:build !freebsd
// +build !freebsd

package osauth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			hash:     "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN.", //nolint:gosec
			password: "123",
			want:     true,
		},
		{
			name:     "sha512 incorrect",
			hash:     "$6$CMWxpgkq.ZosUW8N$gN/MkheCdS9SsPrFS6oOd/k.TMvY2KHztJE5pDMRdN35zr00dyxQr3pYGM4rtPPduUIrEFCwuB7oVgzDbiMfN.", //nolint:gosec
			password: "test",
			want:     false,
		},
		{
			name:     "md5 correct",
			hash:     "$1$YW4a91HG$31CtH9bzW/oyJ1VOD.H/d/", //nolint:gosec
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
	// NOTE: This test only ensures the yescrypt branch executes without panicking.
	// Avoid asserting true/false because yescrypt parameters may vary across
	// environments and producing a deterministic yescrypt hash in tests is
	// environment-dependent.
	yesHash := "$y$e0801$w1Jl9GJH1j4h0w==$Wj2b7m2vWw2m3l1iQe8qvQ=="
	_ = VerifyPasswordHash(yesHash, "password")
}

// nolint:gosec
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
	assert.NoError(t, err)
	assert.Equal(t, 8, len(users))

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

func TestParseIntString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "whitespace only",
			input: "   ",
			want:  0,
		},
		{
			name:  "valid integer",
			input: "42",
			want:  42,
		},
		{
			name:  "valid with surrounding spaces",
			input: "  7  ",
			want:  7,
		},
		{
			name:  "negative integer",
			input: "-3",
			want:  -3,
		},
		{
			name:  "plus sign",
			input: "+5",
			want:  5,
		},
		{
			name:  "non-numeric",
			input: "abc",
			want:  0,
		},
		{
			name:  "mixed numeric and alpha",
			input: "12abc",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseIntString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
