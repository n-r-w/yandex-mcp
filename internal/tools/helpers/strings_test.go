package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTrimStringFields trims in-place string arguments so handlers can normalize once at the boundary.
func TestTrimStringFields(t *testing.T) {
	t.Parallel()

	first := " alpha "
	second := "\tbeta\n"

	TrimStringFields(&first, &second, nil)

	assert.Equal(t, "alpha", first)
	assert.Equal(t, "beta", second)
}

// TestTrimStrings drops whitespace-only slice members after trimming.
func TestTrimStrings(t *testing.T) {
	t.Parallel()

	result := TrimStrings([]string{" alpha ", "   ", "\tbeta\n"})

	assert.Equal(t, []string{"alpha", "beta"}, result)
	assert.Nil(t, TrimStrings(nil))
}
