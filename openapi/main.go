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

	fmt.Printf("info: generating OpenAPI server from %s (edition: %s)\n", specPath, edition) //nolint:forbidigo // this is the generator's own progress output

	if err := exec.Command("redocly", "bundle", specPath, "-o", "static/openapi.json").Run(); err != nil { //nolint:noctx,gosec
		log.Fatalf("error: failed to bundle the openapi spec: %v", err)
	}

	customerAPI := string(edition) + "-customer@v1"
	if err := exec.Command("redocly", "bundle", customerAPI, "-o", "static/customer.json").Run(); err != nil { //nolint:noctx,gosec
		log.Printf("warning: failed to bundle the customer OpenAPI preview (%s): %v", customerAPI, err)
	}

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", fileServer))

	log.Printf("info: OpenAPI server started (edition: %s)\n", edition)

	if err := http.ListenAndServe(":8080", mux); err != nil { //nolint:gosec
		log.Fatal(err)
	}
}
