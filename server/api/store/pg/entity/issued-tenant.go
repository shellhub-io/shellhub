package entity

import (
	"time"

	"github.com/uptrace/bun"
)

// IssuedTenant records a tenant ID the store assigned to a namespace. The table is append-only and
// every namespace references its row, so a tenant ID issued since migration 057 outlives the
// namespace it named and is never assigned to another one. IDs of namespaces deleted before 057
// were never recorded and can be issued again.
type IssuedTenant struct {
	bun.BaseModel `bun:"table:issued_tenants"`

	ID        string    `bun:"id,pk,type:uuid"`
	CreatedAt time.Time `bun:"created_at"`
}
