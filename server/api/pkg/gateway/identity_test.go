package gateway

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWriteToClearsWhatTheClientSent(t *testing.T) {
	spoofed := http.Header{}
	for _, name := range identityHeaders {
		spoofed.Set(name, "forged")
	}

	identity := new(Identity)
	identity.APIKeyID = uuid.Generate()
	identity.WriteTo(spoofed)

	assert.Equal(t, identity.APIKeyID, spoofed.Get("X-API-Key-ID"))

	for _, name := range identityHeaders {
		if name == "X-API-Key-ID" {
			continue
		}

		assert.Empty(t, spoofed.Get(name), name+" was forged by the client and survived")
	}
}

func TestWriteToClearsEverythingForNoIdentity(t *testing.T) {
	spoofed := http.Header{}
	for _, name := range identityHeaders {
		spoofed.Set(name, "forged")
	}

	var identity *Identity
	identity.WriteTo(spoofed)

	for _, name := range identityHeaders {
		assert.Empty(t, spoofed.Get(name), name+" reached an anonymous handler")
	}
}
