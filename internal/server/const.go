//nolint:lll
package server

const (
	serverName    = "yandex-mcp"	
	setverTitle   = "Yandex MCP Server"

	systemPrompt = `This MCP server provides access to various tools for interacting with Yandex services.
YANDEX WIKI rules:
- Any pages of the *wiki.yandex.* type must be loaded via Yandex Wiki tools. Example: https://wiki.yandex.com/homepage/xxx/ -> wiki_page_get(slug: homepage/xxx)
- Manage the wiki_page_get->fields parameter to retrieve the desired data.

YANDEX TRACKER rules:
- Any pages of the *tracker.yandex.* type must be loaded via Yandex Tracker tools. Example: https://tracker.yandex.ru/CP-269 -> tracker_issue_get(issue_id_or_key: CP-269)`
)
