//go:build docker
// +build docker

package main

import (
	"os"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
)

var AgentPlatform string

func init() {
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		AgentPlatform = "bundle"
	} else {
		AgentPlatform = "docker"
	}

	osauth.DefaultShadowFilename = "/host/etc/shadow"
	sysinfo.DefaultOSReleaseFilename = "/host/etc/os-release"
}
