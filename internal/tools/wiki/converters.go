package wiki

import (
	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// Mapping functions from domain models to tool outputs.

func mapPageToOutput(p *domain.WikiPage) *pageOutputDTO {
	if p == nil {
		return nil
	}

	out := &pageOutputDTO{
		ID:         p.ID,
		PageType:   p.PageType,
		Slug:       p.Slug,
		Title:      p.Title,
		Content:    p.Content,
		Attributes: nil,
		Redirect:   nil,
	}

	if p.Attributes != nil {
		out.Attributes = &attributesOutputDTO{
			CommentsCount:   p.Attributes.CommentsCount,
			CommentsEnabled: p.Attributes.CommentsEnabled,
			CreatedAt:       p.Attributes.CreatedAt,
			IsReadonly:      p.Attributes.IsReadonly,
			Lang:            p.Attributes.Lang,
			ModifiedAt:      p.Attributes.ModifiedAt,
			IsCollaborative: p.Attributes.IsCollaborative,
			IsDraft:         p.Attributes.IsDraft,
		}
	}

	if p.Redirect != nil {
		out.Redirect = &redirectOutputDTO{
			PageID: p.Redirect.PageID,
			Slug:   p.Redirect.Slug,
		}
	}

	return out
}

func mapResourcesPageToOutput(rp *domain.WikiResourcesPage) *resourcesListOutputDTO {
	if rp == nil {
		return nil
	}

	resources := make([]resourceOutputDTO, len(rp.Resources))
	for i, r := range rp.Resources {
		resources[i] = mapResourceToOutput(r)
	}

	return &resourcesListOutputDTO{
		Resources:  resources,
		NextCursor: rp.NextCursor,
		PrevCursor: rp.PrevCursor,
	}
}

func mapResourceToOutput(r domain.WikiResource) resourceOutputDTO {
	out := resourceOutputDTO{
		Type: r.Type,
		Item: nil,
	}

	switch {
	case r.Attachment != nil:
		out.Item = attachmentOutputDTO{
			ID:          r.Attachment.ID,
			Name:        r.Attachment.Name,
			Size:        r.Attachment.Size,
			Mimetype:    r.Attachment.MIMEType,
			DownloadURL: r.Attachment.DownloadURL,
			CreatedAt:   r.Attachment.CreatedAt,
			HasPreview:  r.Attachment.HasPreview,
		}
	case r.Sharepoint != nil:
		out.Item = sharepointResourceOutputDTO{
			ID:        r.Sharepoint.ID,
			Title:     r.Sharepoint.Title,
			Doctype:   r.Sharepoint.Doctype,
			CreatedAt: r.Sharepoint.CreatedAt,
		}
	case r.Grid != nil:
		out.Item = gridResourceOutputDTO{
			ID:        r.Grid.ID,
			Title:     r.Grid.Title,
			CreatedAt: r.Grid.CreatedAt,
		}
	}

	return out
}

func mapGridsPageToOutput(gp *domain.WikiGridsPage) *gridsListOutputDTO {
	if gp == nil {
		return nil
	}

	grids := make([]gridSummaryOutputDTO, len(gp.Grids))
	for i, g := range gp.Grids {
		grids[i] = gridSummaryOutputDTO{
			ID:        g.ID,
			Title:     g.Title,
			CreatedAt: g.CreatedAt,
		}
	}

	return &gridsListOutputDTO{
		Grids:      grids,
		NextCursor: gp.NextCursor,
		PrevCursor: gp.PrevCursor,
	}
}

func mapGridToOutput(g *domain.WikiGrid) *gridOutputDTO {
	if g == nil {
		return nil
	}

	out := &gridOutputDTO{
		ID:          g.ID,
		Title:       g.Title,
		Structure:   nil,
		Rows:        nil,
		Revision:    g.Revision,
		CreatedAt:   g.CreatedAt,
		RichTextFmt: g.RichTextFormat,
		Attributes:  nil,
	}

	if g.Attributes != nil {
		out.Attributes = &attributesOutputDTO{
			CommentsCount:   g.Attributes.CommentsCount,
			CommentsEnabled: g.Attributes.CommentsEnabled,
			CreatedAt:       g.Attributes.CreatedAt,
			IsReadonly:      g.Attributes.IsReadonly,
			Lang:            g.Attributes.Lang,
			ModifiedAt:      g.Attributes.ModifiedAt,
			IsCollaborative: g.Attributes.IsCollaborative,
			IsDraft:         g.Attributes.IsDraft,
		}
	}

	if len(g.Structure) > 0 {
		out.Structure = make([]columnOutputDTO, len(g.Structure))
		for i, c := range g.Structure {
			out.Structure[i] = columnOutputDTO{
				Slug:  c.Slug,
				Title: c.Title,
				Type:  c.Type,
			}
		}
	}

	if len(g.Rows) > 0 {
		out.Rows = make([]gridRowOutputDTO, len(g.Rows))
		for i, r := range g.Rows {
			out.Rows[i] = mapGridRowToOutput(r)
		}
	}

	return out
}

func mapGridRowToOutput(r domain.WikiGridRow) gridRowOutputDTO {
	cells := make(map[string]any, len(r.Cells))
	for k, v := range r.Cells {
		cells[k] = v.Value
	}
	return gridRowOutputDTO{
		ID:    r.ID,
		Cells: cells,
	}
}
