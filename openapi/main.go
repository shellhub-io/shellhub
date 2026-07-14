package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/shellhub-io/shellhub/pkg/envs"
)

func main() {
	edition := envs.CurrentEdition()

	var specPath string

	switch edition {
	case envs.Cloud:
		specPath = "spec/cloud-openapi.yaml"
	case envs.Enterprise:
		specPath = "spec/enterprise-openapi.yaml"
	default:
		specPath = "spec/community-openapi.yaml"
	}

	fmt.Printf("info: generating OpenAPI server from %s (edition: %s)\n", specPath, edition)

	if err := exec.Command("redocly", "bundle", specPath, "-o", "static/openapi.json").Run(); err != nil { //nolint:gosec
		log.Fatalf("error: failed to bundle the openapi spec: %v", err)
	}

	// Also bundle the customer-facing (filtered) spec so it can be previewed at
	// /openapi/customer.html. This applies the drop-non-customer decorator,
	// leaving only the namespace-scoped, api-key usable surface. A failure here
	// must not stop the server: the full spec above is what the frontend codegen
	// and the response validator depend on.
	customerAPI := string(edition) + "-customer@v1"
	if err := exec.Command("redocly", "bundle", customerAPI, "-o", "static/customer.json").Run(); err != nil { //nolint:gosec
		log.Printf("warning: failed to bundle the customer OpenAPI preview (%s): %v", customerAPI, err)
	}

	mux := http.NewServeMux()

	// NOTE: Gateway proxy to serve the OpenAPI spec and the Redoc UI. directly on the /openapi path.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", fileServer))

	log.Printf("info: OpenAPI server started (edition: %s)\n", edition)

	if err := http.ListenAndServe(":8080", mux); err != nil { //nolint:gosec
		log.Fatal(err)
	}
}
