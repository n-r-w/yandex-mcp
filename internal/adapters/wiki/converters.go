package wiki

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// Resource type constants for polymorphic conversion.
const (
	resourceTypeAttachment         = "attachment"
	resourceTypeSharepointResource = "sharepoint_resource"
	resourceTypeGrid               = "grid"
)

func pageToWikiPage(p *pageDTO) *domain.WikiPage {
	if p == nil {
		return nil
	}
	return &domain.WikiPage{
		ID:         p.ID.String(),
		PageType:   p.PageType,
		Slug:       p.Slug,
		Title:      p.Title,
		Content:    p.Content,
		Attributes: attributesToWikiAttributes(p.Attributes),
		Redirect:   redirectToWikiRedirect(p.Redirect),
	}
}

func attributesToWikiAttributes(a *attributesDTO) *domain.WikiAttributes {
	if a == nil {
		return nil
	}
	return &domain.WikiAttributes{
		CommentsCount:   a.CommentsCount,
		CommentsEnabled: a.CommentsEnabled,
		CreatedAt:       a.CreatedAt,
		IsReadonly:      a.IsReadonly,
		Lang:            a.Lang,
		ModifiedAt:      a.ModifiedAt,
		IsCollaborative: a.IsCollaborative,
		IsDraft:         a.IsDraft,
	}
}

func redirectToWikiRedirect(r *redirectDTO) *domain.WikiRedirect {
	if r == nil {
		return nil
	}
	return &domain.WikiRedirect{
		PageID: r.PageID.String(),
		Slug:   r.Slug,
	}
}

func resourcesPageToWikiResourcesPage(rp *resourcesPageDTO) (*domain.WikiResourcesPage, error) {
	if rp == nil {
		return nil, nil //nolint:nilnil // nil input returns nil output by design
	}
	resources := make([]domain.WikiResource, 0, len(rp.Resources))
	for i := range rp.Resources {
		r, err := resourceToWikiResource(&rp.Resources[i])
		if err != nil {
			return nil, fmt.Errorf("resource at index %d: %w", i, err)
		}
		resources = append(resources, *r)
	}
	return &domain.WikiResourcesPage{
		Resources:  resources,
		NextCursor: rp.NextCursor,
		PrevCursor: rp.PrevCursor,
	}, nil
}

func resourceToWikiResource(r *resourceDTO) (*domain.WikiResource, error) {
	if r == nil {
		return nil, nil //nolint:nilnil // nil input returns nil output by design
	}
	result := &domain.WikiResource{
		Type:       r.Type,
		Attachment: nil,
		Sharepoint: nil,
		Grid:       nil,
	}
	if r.Item == nil {
		return result, nil
	}
	switch r.Type {
	case resourceTypeAttachment:
		att, err := convertItemToAttachment(r.Item)
		if err != nil {
			return nil, fmt.Errorf("convert attachment item: %w", err)
		}
		result.Attachment = att
	case resourceTypeSharepointResource:
		sp, err := convertItemToSharepoint(r.Item)
		if err != nil {
			return nil, fmt.Errorf("convert sharepoint item: %w", err)
		}
		result.Sharepoint = sp
	case resourceTypeGrid:
		g, err := convertItemToGrid(r.Item)
		if err != nil {
			return nil, fmt.Errorf("convert grid item: %w", err)
		}
		result.Grid = g
	default:
		// Unknown type: leave all pointers nil
	}
	return result, nil
}

func convertItemToAttachment(item any) (*domain.WikiAttachment, error) {
	// Re-marshal and unmarshal through the typed DTO to leverage json tags
	data, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item: %w", err)
	}
	var att attachmentDTO
	if err := json.Unmarshal(data, &att); err != nil {
		return nil, fmt.Errorf("unmarshal attachment: %w", err)
	}
	return &domain.WikiAttachment{
		ID:          att.ID.String(),
		Name:        att.Name,
		Size:        att.Size,
		MIMEType:    att.Mimetype,
		DownloadURL: att.DownloadURL,
		CreatedAt:   att.CreatedAt,
		HasPreview:  att.HasPreview,
	}, nil
}

func convertItemToSharepoint(item any) (*domain.WikiSharepointResource, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item: %w", err)
	}
	var sp sharepointResourceDTO
	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, fmt.Errorf("unmarshal sharepoint: %w", err)
	}
	return &domain.WikiSharepointResource{
		ID:        sp.ID.String(),
		Title:     sp.Title,
		Doctype:   sp.Doctype,
		CreatedAt: sp.CreatedAt,
	}, nil
}

func convertItemToGrid(item any) (*domain.WikiGridResource, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("marshal item: %w", err)
	}
	var g pageGridSummaryDTO
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("unmarshal grid: %w", err)
	}
	return &domain.WikiGridResource{
		ID:        g.ID.String(),
		Title:     g.Title,
		CreatedAt: g.CreatedAt,
	}, nil
}

func gridsPageToWikiGridsPage(gp *gridsPageDTO) *domain.WikiGridsPage {
	if gp == nil {
		return nil
	}
	grids := make([]domain.WikiGridSummary, 0, len(gp.Grids))
	for i := range gp.Grids {
		grids = append(grids, gridSummaryToWikiGridSummary(&gp.Grids[i]))
	}
	return &domain.WikiGridsPage{
		Grids:      grids,
		NextCursor: gp.NextCursor,
		PrevCursor: gp.PrevCursor,
	}
}

func gridSummaryToWikiGridSummary(gs *pageGridSummaryDTO) domain.WikiGridSummary {
	if gs == nil {
		return domain.WikiGridSummary{
			ID:        "",
			Title:     "",
			CreatedAt: "",
		}
	}
	return domain.WikiGridSummary{
		ID:        gs.ID.String(),
		Title:     gs.Title,
		CreatedAt: gs.CreatedAt,
	}
}

func gridToWikiGrid(g *gridDTO) *domain.WikiGrid {
	if g == nil {
		return nil
	}
	structure := make([]domain.WikiColumn, 0, len(g.Structure))
	for i := range g.Structure {
		structure = append(structure, columnToWikiColumn(&g.Structure[i]))
	}
	rows := make([]domain.WikiGridRow, 0, len(g.Rows))
	for i := range g.Rows {
		rows = append(rows, gridRowToWikiGridRow(&g.Rows[i]))
	}
	return &domain.WikiGrid{
		ID:             g.ID.String(),
		Title:          g.Title,
		Structure:      structure,
		Rows:           rows,
		Revision:       g.Revision,
		CreatedAt:      g.CreatedAt,
		RichTextFormat: g.RichTextFormat,
		Attributes:     attributesToWikiAttributes(g.Attributes),
	}
}

func columnToWikiColumn(c *columnDTO) domain.WikiColumn {
	if c == nil {
		return domain.WikiColumn{
			Slug:  "",
			Title: "",
			Type:  "",
		}
	}
	return domain.WikiColumn{
		Slug:  c.Slug,
		Title: c.Title,
		Type:  c.Type,
	}
}

func gridRowToWikiGridRow(r *gridRowDTO) domain.WikiGridRow {
	if r == nil {
		return domain.WikiGridRow{
			ID:    "",
			Cells: nil,
		}
	}
	cells := make(map[string]domain.WikiGridCell, len(r.Cells))
	for colSlug, value := range r.Cells {
		cells[colSlug] = domain.WikiGridCell{
			Value: cellValueToString(value),
		}
	}
	return domain.WikiGridRow{
		ID:    r.ID.String(),
		Cells: cells,
	}
}

func cellValueToString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// JSON numbers are float64; format without trailing zeros if integer
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return fmt.Sprintf("%v", val)
	case bool:
		return strconv.FormatBool(val)
	default:
		// For complex types (maps, slices), marshal to JSON string
		data, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(data)
	}
}
