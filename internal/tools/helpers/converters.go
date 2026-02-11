package helpers

import (
	"context"
	"fmt"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// ConvertFilterToStringMap converts a map[string]any filter to map[string]string.
func ConvertFilterToStringMap(
	_ context.Context, filter map[string]any, _ domain.Service,
) (map[string]string, error) {
	if filter == nil {
		return nil, nil //nolint:nilnil // nil filter means no filter, not an error
	}
	result := make(map[string]string, len(filter))
	for k, v := range filter {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("filter value for key %q must be a string, got %T", k, v)
		}
		result[k] = s
	}
	return result, nil
}
