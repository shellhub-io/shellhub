package agent

import "log"

func ExampleNewAgentWithConfig() {
	// Creates the agent configuration with the minimum required fields.
	cfg := Config{
		ServerAddress: "http://localhost:80",
		TenantID:      "00000000-0000-4000-0000-000000000000",
		PrivateKey:    "./shellhub.key",
	}

	ag, err := NewAgentWithConfig(&cfg)
	if err != nil {
		panic(err)
	}

	// Initializes agent, generating device identity, loading device information, generating private key,
	// reading public key, probing server information and authorizing device on ShellHub server.
	if err := ag.Initialize(); err != nil {
		panic(err)
	}

	listing := make(chan bool)
	go func() {
		<-listing

		log.Println("listing")
	}()

	ag.Listen(listing) //nolint:errcheck
}

func ExampleNewAgent() {
	ag, err := NewAgent("http://localhost:80", "00000000-0000-4000-0000-000000000000", "./shellhub.key")
	if err != nil {
		panic(err)
	}

	// Initializes agent, generating device identity, loading device information, generating private key,
	// reading public key, probing server information and authorizing device on ShellHub server.
	if err := ag.Initialize(); err != nil {
		panic(err)
	}

	listing := make(chan bool)
	go func() {
		<-listing

		log.Println("listing")
	}()

	ag.Listen(listing) //nolint:errcheck
}
