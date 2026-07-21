// Package contract is the users bounded context's public port. Other modules
// depend ONLY on this package (interface + plain structs), never on the
// module's internals. When users becomes its own service, an HTTP/gRPC client
// implementing UserDirectory replaces the in-process implementation.
package contract

import (
	"context"
	"time"
)

// UserSummary is the read-only projection of a user other contexts may see.
type UserSummary struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

// UserDirectory exposes user lookups to other bounded contexts.
type UserDirectory interface {
	// FindByIDs returns summaries for the given ids (missing ids are omitted).
	FindByIDs(ctx context.Context, ids []int64) ([]UserSummary, error)
}
