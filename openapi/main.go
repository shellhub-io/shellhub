package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/shellhub-io/shellhub/pkg/envs"
)

const (
	// CommunitySpecPath is the path to the Community edition OpenAPI spec.
	CommunitySpecPath = "spec/community-openapi.yaml"
	// EnterpriseSpecPath is the path to the Enterprise edition OpenAPI spec.
	EnterpriseSpecPath = "spec/enterprise-openapi.yaml"
	// CloudSpecPath is the path to the Cloud edition OpenAPI spec.
	CloudSpecPath = "spec/cloud-openapi.yaml"
)

func main() {
	var path string

	// NOTE: Determine which OpenAPI spec to use based on the edition.
	// Currently, the path are hardcoded, but in the future, this could be
	// made configurable via command-line flags or environment variables.
	switch {
	case envs.IsCommunity():
		path = CommunitySpecPath
		log.Printf("info: Running in Community edition")
	// WARN: The order of these checks matters. Cloud should be checked before Enterprise
	// because Cloud is a superset of Enterprise.
	case envs.IsCloud():
		path = CloudSpecPath
		log.Printf("info: Running in Cloud edition")
	case envs.IsEnterprise():
		path = EnterpriseSpecPath
		log.Printf("info: Running in Enterprise edition")
	}

	fmt.Printf("info: generating OpenAPI server from %s\n", path)

	if err := exec.Command("redocly", "bundle", path, "-o", "static/openapi.json").Run(); err != nil { //nolint:gosec
		log.Fatalf("error: failed to bundle the openapi spec: %v", err)
	}

	mux := http.NewServeMux()

	// NOTE: Gateway proxy to serve the OpenAPI spec and the Redoc UI. directly on the /openapi path.
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", fileServer))

	log.Printf("info: OpenAPI server started\n")

	if err := http.ListenAndServe(":8080", mux); err != nil { //nolint:gosec
		log.Fatal(err)
	}
}
