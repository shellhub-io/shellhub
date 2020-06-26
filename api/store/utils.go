package store

import "context"

type Tenant struct {
	ID string
}

func TenantFromContext(ctx context.Context) *Tenant {
	if v, ok := ctx.Value("tenant").(string); ok {
		return &Tenant{v}
	}

	return nil
}
