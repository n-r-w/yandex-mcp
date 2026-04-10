package domain

// WikiListResourcesOpts represents options for listing Wiki page resources.
type WikiListResourcesOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
	Query          string
	Types          string
}

// WikiListGridsOpts represents options for listing Wiki page grids.
type WikiListGridsOpts struct {
	Cursor         string
	PageSize       int
	OrderBy        string
	OrderDirection string
}

// WikiGetGridOpts represents options for getting a specific Wiki grid.
type WikiGetGridOpts struct {
	Fields   []string
	Filter   string
	OnlyCols string
	OnlyRows string
	Revision string
	Sort     string
}

// WikiListDescendantsOpts represents options for listing Wiki page descendants.
type WikiListDescendantsOpts struct {
	Actuality string
	Cursor    string
	PageSize  int
}

// WikiGetPageOpts represents options for getting a Wiki page.
type WikiGetPageOpts struct {
	Fields          []string
	RevisionID      string
	RaiseOnRedirect bool
}
