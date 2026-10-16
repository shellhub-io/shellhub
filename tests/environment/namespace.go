package environment

import "context"

// CreateNamespace runs the admin CLI's namespace create, as [DockerCompose.NewNamespace] does, and
// returns its refusal instead of failing the test. ctx
// bounds the exec. It returns the error of an exec that could not run, and an error carrying the
// CLI's output when the CLI exits non-zero.
func (dc *DockerCompose) CreateNamespace(ctx context.Context, owner, name, tenant string) error {
	return dc.stack.NewNamespace(ctx, owner, name, tenant)
}
