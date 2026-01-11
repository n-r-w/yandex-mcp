package domain

// WikiTool represents a specific Wiki tool that can be registered.
type WikiTool int

// WikiTool constants represent individual Wiki tools.
const (
	WikiToolPageGetBySlug WikiTool = iota
	WikiToolPageGetByID
	WikiToolResourcesList
	WikiToolGridsList
	WikiToolGridGet
	WikiToolCount // used to verify list completeness
)

// String returns the MCP tool name for the WikiTool.
func (w WikiTool) String() string {
	switch w {
	case WikiToolPageGetBySlug:
		return "wiki_page_get"
	case WikiToolPageGetByID:
		return "wiki_page_get_by_id"
	case WikiToolResourcesList:
		return "wiki_page_resources_list"
	case WikiToolGridsList:
		return "wiki_page_grids_list"
	case WikiToolGridGet:
		return "wiki_grid_get"
	case WikiToolCount:
		return ""
	default:
		return ""
	}
}

// WikiAllTools returns all wiki tools in stable order.
func WikiAllTools() []WikiTool {
	return []WikiTool{
		WikiToolPageGetBySlug,
		WikiToolPageGetByID,
		WikiToolResourcesList,
		WikiToolGridsList,
		WikiToolGridGet,
	}
}
