//nolint:lll // systemPrompt contains long URLs that must stay on single line for readability
package server

const (
	serverName  = "yandex-mcp"
	serverTitle = "Yandex MCP Server"

	baseSystemPrompt = `This MCP server provides access to various tools for interacting with Yandex services.
YANDEX WIKI rules:
- Any pages of the *wiki.yandex.* type must be loaded via Yandex Wiki tools. Example: https://wiki.yandex.com/homepage/xxx/ -> wiki_page_get(slug: homepage/xxx)
- Manage the wiki_page_get->fields parameter to retrieve the desired data.

YANDEX TRACKER rules:
- Any pages of the *tracker.yandex.* type must be loaded via Yandex Tracker tools. Example: https://tracker.yandex.ru/CP-269 -> tracker_issue_get(issue_id_or_key: CP-269)`

	writeToolsWarning = `

WRITE TOOLS are enabled on this server (e.g., tracker_issue_create).
These tools MUST only be used when the user has given explicit permission to perform the write operation.
Always confirm with the user before creating, modifying, or deleting any data.`
)

// BuildSystemPrompt returns the system prompt, appending write-tools warning when enabled.
func BuildSystemPrompt(writeToolsEnabled bool) string {
	if writeToolsEnabled {
		return baseSystemPrompt + writeToolsWarning
	}
	return baseSystemPrompt
}
