// +build docker

package main

import (
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
)

var AgentPlatform = "docker"

func init() {
	osauth.DefaultShadowFilename = "/host/etc/shadow"
	sysinfo.DefaultOSReleaseFilename = "/host/etc/os-release"
}
