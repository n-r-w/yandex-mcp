package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt_WriteToolsDisabled(t *testing.T) {
	t.Parallel()

	prompt := BuildSystemPrompt(false)

	assert.Equal(t, baseSystemPrompt, prompt)
	assert.NotContains(t, prompt, "WRITE TOOLS")
}

func TestBuildSystemPrompt_WriteToolsEnabled(t *testing.T) {
	t.Parallel()

	prompt := BuildSystemPrompt(true)

	assert.Contains(t, prompt, baseSystemPrompt)
	assert.Contains(t, prompt, "WRITE TOOLS")
	assert.Contains(t, prompt, "explicit permission")
}
