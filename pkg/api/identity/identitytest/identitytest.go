// Package identitytest stands in for the authentication middleware in tests that drive a router
// built without it.
package identitytest

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/identity"
)

// As returns req carrying caller as its authenticated identity.
func As(req *http.Request, caller identity.Identity) *http.Request {
	return req.WithContext(identity.NewContext(req.Context(), caller))
}

// Serve returns next serving every request as caller, for a test that reaches the handler over a
// real connection, where no context crosses from the client.
func Serve(next http.Handler, caller identity.Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, As(req, caller))
	})
}
