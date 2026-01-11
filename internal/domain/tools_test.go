package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

func TestWikiToolLists_Complete(t *testing.T) {
	t.Parallel()

	// WikiAllTools length should equal WikiToolCount
	assert.Len(t, domain.WikiAllTools(), int(domain.WikiToolCount), "WikiAllTools should contain all wiki tools")

	// Check for duplicates in WikiAllTools
	seen := make(map[domain.WikiTool]bool)
	for _, tool := range domain.WikiAllTools() {
		assert.False(t, seen[tool], "WikiAllTools should not contain duplicates")
		seen[tool] = true
	}

}

func TestTrackerToolLists_Complete(t *testing.T) {
	t.Parallel()

	// TrackerAllTools length should equal TrackerToolCount
	assert.Len(t, domain.TrackerAllTools(), int(domain.TrackerToolCount), "TrackerAllTools should contain all tracker tools")

	// Check for duplicates in TrackerAllTools
	seen := make(map[domain.TrackerTool]bool)
	for _, tool := range domain.TrackerAllTools() {
		assert.False(t, seen[tool], "TrackerAllTools should not contain duplicates")
		seen[tool] = true
	}

}

func TestTrackerTool_String_AllToolsHaveNames(t *testing.T) {
	t.Parallel()

	for _, tool := range domain.TrackerAllTools() {
		name := tool.String()
		assert.NotEmpty(t, name, "TrackerTool(%d) should have a non-empty string representation", tool)
		assert.Contains(t, name, "tracker_", "TrackerTool(%d) string should contain 'tracker_' prefix", tool)
	}
}

func TestTrackerTool_String_InvalidValue(t *testing.T) {
	t.Parallel()

	invalidTool := domain.TrackerTool(9999)
	assert.Empty(t, invalidTool.String(), "Invalid TrackerTool should return empty string")
}

func TestWikiTool_String_AllToolsHaveNames(t *testing.T) {
	t.Parallel()

	for _, tool := range domain.WikiAllTools() {
		name := tool.String()
		assert.NotEmpty(t, name, "WikiTool(%d) should have a non-empty string representation", tool)
		assert.Contains(t, name, "wiki_", "WikiTool(%d) string should contain 'wiki_' prefix", tool)
	}
}

func TestWikiTool_String_InvalidValue(t *testing.T) {
	t.Parallel()

	invalidTool := domain.WikiTool(9999)
	assert.Empty(t, invalidTool.String(), "Invalid WikiTool should return empty string")
}
