package domain

// WikiListResourcesOpts represents options for listing Wiki page resources.
type WikiListResourcesOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
	Query          string
	Types          string
	PageIDLegacy   int
}

// WikiListGridsOpts represents options for listing Wiki page grids.
type WikiListGridsOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
	PageIDLegacy   int
}

// WikiGetGridOpts represents options for getting a specific Wiki grid.
type WikiGetGridOpts struct {
	Fields   []string
	Filter   string
	OnlyCols string
	OnlyRows string
	Revision int
	Sort     string
}

// WikiGetPageOpts represents options for getting a Wiki page.
type WikiGetPageOpts struct {
	Fields          []string
	RevisionID      int
	RaiseOnRedirect bool
}
