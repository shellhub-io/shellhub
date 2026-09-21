package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
)

const (
	bundledSpec = "../openapi/static/openapi.json"

	strictModeMismatchMarker = "the response does not match the OpenAPI schema"

	strictModeEnabledMarker = "Enabling OpenAPI response validation in strict mode"

	validationRanMarker = "OpenAPI response validation passed"

	sweepObserver         = "sweep-observer"
	sweepObserverEmail    = "sweep-observer@shellhub.io"
	sweepObserverPassword = "password"

	unresolvablePathParam = "e2e-sweep"
)

var (
	sweepPathParams = regexp.MustCompile(`\{[^}]+\}`)

	pathsTheValidatorSkips = map[string]bool{
		"/healthcheck":      true,
		"/api/healthcheck":  true,
		"/info":             true,
		"/api/info":         true,
		"/install.sh":       true,
		"/kickstart.sh":     true,
		"/api/install":      true,
		"/ssh/connection":   true,
		"/agent/connection": true,
		"/ssh/revdial":      true,
		"/ws/ssh":           true,
	}

	prefixesTheValidatorSkips = []string{"/metrics", "/internal"}

	sweepMethods = map[string]bool{
		http.MethodGet:    true,
		http.MethodPost:   true,
		http.MethodPut:    true,
		http.MethodPatch:  true,
		http.MethodDelete: true,
	}
)

type sweepOperation struct {
	method string
	path   string
}

func validatorSkips(path string) bool {
	if pathsTheValidatorSkips[path] {
		return true
	}

	for _, prefix := range prefixesTheValidatorSkips {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func sweepOperations(t *testing.T) []sweepOperation {
	t.Helper()

	content, err := os.ReadFile(bundledSpec)
	require.NoError(t, err, "reading %s: run the openapi bundle before the suite", bundledSpec)

	doc := struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}{}
	require.NoError(t, json.Unmarshal(content, &doc))
	require.NotEmpty(t, doc.Paths, "the bundled spec declares no paths")

	operations := make([]sweepOperation, 0, len(doc.Paths))

	for path, methods := range doc.Paths {
		if validatorSkips(path) {
			continue
		}

		for method := range methods {
			if !sweepMethods[strings.ToUpper(method)] {
				continue
			}

			operations = append(operations, sweepOperation{method: strings.ToUpper(method), path: path})
		}
	}

	sort.Slice(operations, func(i, j int) bool {
		if operations[i].path != operations[j].path {
			return operations[i].path < operations[j].path
		}

		return operations[i].method < operations[j].method
	})

	return operations
}

// TestOpenAPIDescribesEveryRefusal calls every operation the bundled spec declares without a
// credential and again as an observer. It pins that strict mode is on and that every documented
// operation answers a shape the document allows on the paths the validator reaches.
//
// It does not reach a refusal the error handler renders: the validator returns before validating
// when a handler returns an error, and pkgmiddleware.Log renders outside the capture. The source
// ratchet in server/api/routes is what catches a regression to c.NoContent.
func TestOpenAPIDescribesEveryRefusal(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")
	compose.NewUser(t, sweepObserver, sweepObserverEmail, sweepObserverPassword)
	compose.NewMember(t, sweepObserver, ShellHubNamespaceName, string(authorizer.RoleObserver))

	observer := compose.AuthUser(t, sweepObserver, sweepObserverPassword)

	require.NotEmpty(t, observer.Token)

	compose.AwaitServerLog(t, strictModeEnabledMarker)
	compose.AwaitServerLog(t, validationRanMarker)

	callers := []struct {
		name  string
		token string
	}{
		{name: "anonymous", token: ""},
		{name: "observer", token: observer.Token},
	}

	for _, operation := range sweepOperations(t) {
		path := sweepPathParams.ReplaceAllString(operation.path, unresolvablePathParam)

		for _, caller := range callers {
			t.Run(caller.name+" "+operation.method+" "+operation.path, func(t *testing.T) {
				req := compose.Anonymous(ctx)
				if caller.token != "" {
					req = req.SetAuthToken(caller.token)
				}

				resp, err := req.Execute(operation.method, path)
				require.NoError(t, err)
				require.NotContains(t, string(resp.Body()), strictModeMismatchMarker,
					"%s %s answered a shape the spec does not describe", operation.method, operation.path)
			})
		}
	}

	t.Run("the observer token still authenticates", func(t *testing.T) {
		resp, err := compose.Anonymous(ctx).SetAuthToken(observer.Token).Get("/api/auth/user")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(),
			"a swept operation changed the observer, so every observer pass before this ran unauthenticated")
	})
}
