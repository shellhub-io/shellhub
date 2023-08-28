package agent

func ExampleNewAgentWithConfig() {
	_, err := NewAgentWithConfig(&Config{
		ServerAddress: "http://localhost:80",
		TenantID:      "00000000-0000-4000-0000-000000000000",
		PrivateKey:    "./shellhub.key",
	})
	if err != nil {
		panic(err)
	}
}

func ExampleNewAgent() {
	_, err := NewAgent("http://localhost:80", "00000000-0000-4000-0000-000000000000", "./shellhub.key")
	if err != nil {
		panic(err)
	}
}
