package memo

import "github.com/memocash/index/db/client"

const (
	// DefaultPageSize is the public default for paginated memo collections.
	DefaultPageSize = client.MediumLimit
	// MaxPageSize is the largest explicit page accepted by the GraphQL layer.
	MaxPageSize = client.HugeLimit
)

func NormalizePageLimit(limit uint32) uint32 {
	if limit == 0 {
		return DefaultPageSize
	} else if limit > MaxPageSize {
		return MaxPageSize
	}
	return limit
}
