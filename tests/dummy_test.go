package main

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
)

// TestDummy is a test function that demonstrates how to use the testing integration framework.
func TestDummy(t *testing.T) {
	// Load environment variables for the test
	env := loadEnv(t)

	// Modifying the environment variables for a specific test scenario:
	// env["SOME_CUSTOM_VARIABLE"] = "custom_value"

	// Set up a test environment and execute the core test logic within the callback function
	withTestEnvironment(t, env, func() {
		client := resty.New()

		resp, err := client.R().Get(fmt.Sprintf("http://localhost:%s/info", env["SHELLHUB_HTTP_PORT"]))

		fmt.Println(err)
		fmt.Println(resp)
	})
}
