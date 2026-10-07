package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestLegacyPasswordDigest covers an account whose password is still stored as an unsalted SHA256
// digest, as accounts created before bcrypt were: the right password logs in and the login rehashes
// it with bcrypt, and a wrong one leaves the digest as it was.
func TestLegacyPasswordDigest(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	const username = "legacy"

	compose.NewUser(t, username, username+"@ossystems.com.br", ShellHubPassword)

	sum := sha256.Sum256([]byte(ShellHubPassword))
	legacy := hex.EncodeToString(sum[:])
	compose.SetUserPasswordDigest(t, username, legacy)

	login := func(t *testing.T, password string) int {
		t.Helper()

		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"username": username, "password": password}).
			Post("/api/login")
		require.NoError(t, err)

		return resp.StatusCode()
	}

	t.Run("a wrong password is refused and leaves the digest alone", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, login(t, "not-"+ShellHubPassword))
		assert.Equal(t, legacy, compose.UserPasswordDigest(t, username))
	})

	t.Run("the right password logs in and rehashes the digest with bcrypt", func(t *testing.T) {
		require.Equal(t, http.StatusOK, login(t, ShellHubPassword))

		upgraded := compose.UserPasswordDigest(t, username)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(upgraded), []byte(ShellHubPassword)),
			"the stored digest should be a bcrypt hash of the password")

		assert.Equal(t, http.StatusOK, login(t, ShellHubPassword), "the password should keep working against the bcrypt digest")
	})
}
