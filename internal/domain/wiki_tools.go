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
	WikiToolPageCreate
	WikiToolPageUpdate
	WikiToolPageAppend
	WikiToolGridCreate
	WikiToolGridCellsUpdate
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
	case WikiToolPageCreate:
		return "wiki_page_create"
	case WikiToolPageUpdate:
		return "wiki_page_update"
	case WikiToolPageAppend:
		return "wiki_page_append_content"
	case WikiToolGridCreate:
		return "wiki_grid_create"
	case WikiToolGridCellsUpdate:
		return "wiki_grid_update_cells"
	case WikiToolCount:
		return ""
	default:
		return ""
	}
}

// WikiReadOnlyTools returns the default read-only tools.
func WikiReadOnlyTools() []WikiTool {
	return []WikiTool{
		WikiToolPageGetBySlug,
		WikiToolPageGetByID,
		WikiToolResourcesList,
		WikiToolGridsList,
		WikiToolGridGet,
	}
}

// WikiWriteTools returns the write tools enabled via --wiki-write flag.
func WikiWriteTools() []WikiTool {
	return []WikiTool{
		WikiToolPageCreate,
		WikiToolPageUpdate,
		WikiToolPageAppend,
		WikiToolGridCreate,
		WikiToolGridCellsUpdate,
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
		WikiToolPageCreate,
		WikiToolPageUpdate,
		WikiToolPageAppend,
		WikiToolGridCreate,
		WikiToolGridCellsUpdate,
	}
}
