package environment

import (
	"context"
	"os"
	"os/exec"
)

// BuildAgentImage builds agent:test, the agent the console e2e specs start per test, with the
// root/password login those specs sign in with. repoRoot is the build context and must hold
// agent/Dockerfile.test. The build output goes to stderr.
func BuildAgentImage(ctx context.Context, repoRoot string) error {
	cmd := exec.CommandContext(ctx,
		"docker", "build",
		"--tag", "agent:test",
		"--file", "agent/Dockerfile.test",
		"--build-arg", "USERNAME=root",
		"--build-arg", "PASSWORD=password",
		".",
	)
	cmd.Dir = repoRoot
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
