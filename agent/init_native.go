//go:build !docker
// +build !docker

package main

// AgentVersion store the version to be embed inside the binary. This is
// injected using `-ldflags` build option.
//
//	go build -ldflags "-X main.AgentVersion=1.2.3"
//
// If set to `latest`, the auto-updating mechanism is disabled. This is intended
// to be used during development only.
var AgentVersion string

// AgentPlatform stores what platform the agent is running on. This is injected in build time in the [ShellHub Agent]
// implementation.
//
// [ShellHub Agent]: https://github.com/shellhub-io/shellhub/tree/master/agent
var AgentPlatform string

func init() {
	AgentPlatform = "native"
}
