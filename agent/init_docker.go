//go:build docker
// +build docker

package main

import (
	"os"

	"github.com/shellhub-io/shellhub/agent/sysinfo"
)

var AgentPlatform string

func init() {
	if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
		AgentPlatform = "bundle"
	} else {
		AgentPlatform = "docker"
	}

	sysinfo.DefaultOSReleaseFilename = "/host/etc/os-release"
}
