package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	path := "spec/openapi.yaml"

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
