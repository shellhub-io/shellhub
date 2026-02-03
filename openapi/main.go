package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/shellhub-io/shellhub/pkg/envs"
)

func main() {
	// Determine which spec to use based on SHELLHUB_CLOUD and SHELLHUB_ENTERPRISE
	isCloud := envs.IsCloud()
	isEnterprise := envs.IsEnterprise()

	var specPath string
	var edition string

	switch {
	case isCloud:
		specPath = "spec/cloud-openapi.yaml"
		edition = "cloud"
	case isEnterprise:
		specPath = "spec/enterprise-openapi.yaml"
		edition = "enterprise"
	default:
		specPath = "spec/community-openapi.yaml"
		edition = "community"
	}

	fmt.Printf("info: generating OpenAPI server from %s (edition: %s)\n", specPath, edition)

	if err := exec.Command("redocly", "bundle", specPath, "-o", "static/openapi.json").Run(); err != nil { //nolint:gosec
		log.Fatalf("error: failed to bundle the openapi spec: %v", err)
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
