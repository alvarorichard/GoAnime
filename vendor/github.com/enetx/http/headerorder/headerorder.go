// Package headerorder exposes the common wire-header sorter to protocol adapters
// maintained outside the HTTP module. All protocols share one implementation.
package headerorder

import "github.com/enetx/http/internal/headerorder"

const (
	HeaderOrderKey  = headerorder.HeaderOrderKey
	PHeaderOrderKey = headerorder.PHeaderOrderKey
	HeaderCaseKey   = headerorder.HeaderCaseKey
)

type Field = headerorder.Field

// IsMetadataKey identifies private ordering and spelling controls, which must
// be excluded from field validation and encoding by every protocol adapter.
func IsMetadataKey(key string) bool { return headerorder.IsMetadataKey(key) }

func Indices(names []string) map[string]int { return headerorder.Indices(names) }

// Sorted returns independent ordering without modifying headers or exclusions.
func Sorted(h map[string][]string, order map[string]int, exclude map[string]bool) []Field {
	return headerorder.Sorted(h, order, exclude)
}
