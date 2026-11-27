package routes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/require"
)

type deviceFields struct {
	info      *requests.DeviceInfo
	identity  *requests.DeviceIdentity
	publicKey *string
}

type deviceFieldBound struct {
	field string
	limit int
	set   func(f deviceFields, value string)
}

var deviceFieldBounds = []deviceFieldBound{
	{"info.id", 64, func(f deviceFields, v string) { f.info.ID = v }},
	{"info.pretty_name", 256, func(f deviceFields, v string) { f.info.PrettyName = v }},
	{"info.version", 32, func(f deviceFields, v string) { f.info.Version = v }},
	{"info.arch", 16, func(f deviceFields, v string) { f.info.Arch = v }},
	{"info.platform", 32, func(f deviceFields, v string) { f.info.Platform = v }},
	{"identity.mac", 64, func(f deviceFields, v string) { f.identity.MAC = v }},
	{"public_key", 4096, func(f deviceFields, v string) { *f.publicKey = v }},
}

var rsa4096PublicKeyPEM = sync.OnceValues(func() (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)})), nil
})

func rsa4096AgentPublicKey(t *testing.T) string {
	t.Helper()

	key, err := rsa4096PublicKeyPEM()
	require.NoError(t, err)

	return key
}

func atLimitConnectorImage(t *testing.T) string {
	t.Helper()

	image := "registry.example.com/platform/" + strings.Repeat("a", 150) + ":1.0@sha256:" + strings.Repeat("0", 64)
	require.Len(t, image, 256)

	return image
}

func postDeviceJSON(t *testing.T, mock *mocks.MockService, url string, body any) int {
	t.Helper()

	data, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader(string(data)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	newTestRouter(t, mock).ServeHTTP(rec, req)

	return rec.Result().StatusCode
}
