package tracker

import (
	"encoding/base64"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// Mapping functions from domain models to tool outputs.

func mapIssueToOutput(i *domain.TrackerIssue) *issueOutputDTO {
	if i == nil {
		return nil
	}

	return &issueOutputDTO{
		Self:            i.Self,
		ID:              i.ID,
		Key:             i.Key,
		Version:         i.Version,
		Summary:         i.Summary,
		Description:     i.Description,
		StatusStartTime: i.StatusStartTime,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
		ResolvedAt:      i.ResolvedAt,
		Status:          mapStatusToOutput(i.Status),
		Type:            mapTypeToOutput(i.Type),
		Priority:        mapPriorityToOutput(i.Priority),
		Queue:           mapQueueToOutput(i.Queue),
		Assignee:        mapUserToOutput(i.Assignee),
		CreatedBy:       mapUserToOutput(i.CreatedBy),
		UpdatedBy:       mapUserToOutput(i.UpdatedBy),
		Votes:           i.Votes,
		Favorite:        i.Favorite,
	}
}

func mapStatusToOutput(s *domain.TrackerStatus) *statusOutputDTO {
	if s == nil {
		return nil
	}
	return &statusOutputDTO{
		Self:    s.Self,
		ID:      s.ID,
		Key:     s.Key,
		Display: s.Display,
	}
}

func mapTypeToOutput(t *domain.TrackerIssueType) *typeOutputDTO {
	if t == nil {
		return nil
	}
	return &typeOutputDTO{
		Self:    t.Self,
		ID:      t.ID,
		Key:     t.Key,
		Display: t.Display,
	}
}

func mapPriorityToOutput(p *domain.TrackerPriority) *priorityOutputDTO {
	if p == nil {
		return nil
	}
	return &priorityOutputDTO{
		Self:    p.Self,
		ID:      p.ID,
		Key:     p.Key,
		Display: p.Display,
	}
}

func mapQueueToOutput(q *domain.TrackerQueue) *queueOutputDTO {
	if q == nil {
		return nil
	}
	return &queueOutputDTO{
		Self:           q.Self,
		ID:             q.ID,
		Key:            q.Key,
		Display:        q.Display,
		Name:           q.Name,
		Version:        q.Version,
		Lead:           mapUserToOutput(q.Lead),
		AssignAuto:     q.AssignAuto,
		AllowExternals: q.AllowExternals,
		DenyVoting:     q.DenyVoting,
	}
}

func mapUserToOutput(u *domain.TrackerUser) *userOutputDTO {
	if u == nil {
		return nil
	}
	return &userOutputDTO{
		Self:        u.Self,
		ID:          u.ID,
		UID:         u.UID,
		Login:       u.Login,
		Display:     u.Display,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		CloudUID:    u.CloudUID,
		PassportUID: u.PassportUID,
	}
}

func mapTransitionToOutput(t *domain.TrackerTransition) *transitionOutputDTO {
	if t == nil {
		return nil
	}
	return &transitionOutputDTO{
		ID:      t.ID,
		Display: t.Display,
		Self:    t.Self,
		To:      mapStatusToOutput(t.To),
	}
}

func mapCommentToOutput(c *domain.TrackerComment) *commentOutputDTO {
	if c == nil {
		return nil
	}
	return &commentOutputDTO{
		ID:        c.ID,
		LongID:    c.LongID,
		Self:      c.Self,
		Text:      c.Text,
		Version:   c.Version,
		Type:      c.Type,
		Transport: c.Transport,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		CreatedBy: mapUserToOutput(c.CreatedBy),
		UpdatedBy: mapUserToOutput(c.UpdatedBy),
	}
}

func mapSearchResultToOutput(r *domain.TrackerIssuesPage) *searchIssuesOutputDTO {
	if r == nil {
		return nil
	}

	issues := make([]issueOutputDTO, len(r.Issues))
	for i, issue := range r.Issues {
		out := mapIssueToOutput(&issue)
		if out != nil {
			issues[i] = *out
		}
	}

	return &searchIssuesOutputDTO{
		Issues:      issues,
		TotalCount:  r.TotalCount,
		TotalPages:  r.TotalPages,
		ScrollID:    r.ScrollID,
		ScrollToken: r.ScrollToken,
		NextLink:    r.NextLink,
	}
}

func mapTransitionsToOutput(transitions []domain.TrackerTransition) *transitionsListOutputDTO {
	out := make([]transitionOutputDTO, len(transitions))
	for i, t := range transitions {
		mapped := mapTransitionToOutput(&t)
		if mapped != nil {
			out[i] = *mapped
		}
	}
	return &transitionsListOutputDTO{Transitions: out}
}

func mapQueuesResultToOutput(r *domain.TrackerQueuesPage) *queuesListOutputDTO {
	if r == nil {
		return nil
	}

	queues := make([]queueOutputDTO, len(r.Queues))
	for i, q := range r.Queues {
		out := mapQueueToOutput(&q)
		if out != nil {
			queues[i] = *out
		}
	}

	return &queuesListOutputDTO{
		Queues:     queues,
		TotalCount: r.TotalCount,
		TotalPages: r.TotalPages,
	}
}

func mapCommentsResultToOutput(r *domain.TrackerCommentsPage) *commentsListOutputDTO {
	if r == nil {
		return nil
	}

	comments := make([]commentOutputDTO, len(r.Comments))
	for i, c := range r.Comments {
		out := mapCommentToOutput(&c)
		if out != nil {
			comments[i] = *out
		}
	}

	return &commentsListOutputDTO{
		Comments: comments,
		NextLink: r.NextLink,
	}
}

func mapAttachmentToOutput(a *domain.TrackerAttachment) *attachmentOutputDTO {
	if a == nil {
		return nil
	}
	var metadata *attachmentMetadataOutputDTO
	if a.Metadata != nil {
		metadata = &attachmentMetadataOutputDTO{
			Size: a.Metadata.Size,
		}
	}
	return &attachmentOutputDTO{
		ID:           a.ID,
		Name:         a.Name,
		ContentURL:   a.ContentURL,
		ThumbnailURL: a.ThumbnailURL,
		Mimetype:     a.Mimetype,
		Size:         a.Size,
		CreatedAt:    a.CreatedAt,
		CreatedBy:    mapUserToOutput(a.CreatedBy),
		Metadata:     metadata,
	}
}

func mapAttachmentsToOutput(attachments []domain.TrackerAttachment) *attachmentsListOutputDTO {
	result := make([]attachmentOutputDTO, len(attachments))
	for i, a := range attachments {
		out := mapAttachmentToOutput(&a)
		if out != nil {
			result[i] = *out
		}
	}
	return &attachmentsListOutputDTO{
		Attachments: result,
	}
}

func mapAttachmentContentToOutput(content *domain.TrackerAttachmentContent) *attachmentContentOutputDTO {
	if content == nil {
		return nil
	}
	return &attachmentContentOutputDTO{
		FileName:      content.FileName,
		ContentType:   content.ContentType,
		ContentBase64: base64.StdEncoding.EncodeToString(content.Data),
		Size:          int64(len(content.Data)),
	}
}

func mapQueueDetailToOutput(q *domain.TrackerQueueDetail) *queueDetailOutputDTO {
	if q == nil {
		return nil
	}
	return &queueDetailOutputDTO{
		Self:            q.Self,
		ID:              q.ID,
		Key:             q.Key,
		Display:         q.Display,
		Name:            q.Name,
		Description:     q.Description,
		Version:         q.Version,
		Lead:            mapUserToOutput(q.Lead),
		AssignAuto:      q.AssignAuto,
		AllowExternals:  q.AllowExternals,
		DenyVoting:      q.DenyVoting,
		DefaultType:     mapTypeToOutput(q.DefaultType),
		DefaultPriority: mapPriorityToOutput(q.DefaultPriority),
	}
}

func mapUserDetailToOutput(u *domain.TrackerUserDetail) *userDetailOutputDTO {
	if u == nil {
		return nil
	}
	return &userDetailOutputDTO{
		Self:        u.Self,
		ID:          u.ID,
		UID:         u.UID,
		TrackerUID:  u.TrackerUID,
		Login:       u.Login,
		Display:     u.Display,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		CloudUID:    u.CloudUID,
		PassportUID: u.PassportUID,
		HasLicense:  u.HasLicense,
		Dismissed:   u.Dismissed,
		External:    u.External,
	}
}

func mapUsersPageToOutput(p *domain.TrackerUsersPage) *usersListOutputDTO {
	if p == nil {
		return nil
	}
	users := make([]userDetailOutputDTO, len(p.Users))
	for i, u := range p.Users {
		out := mapUserDetailToOutput(&u)
		if out != nil {
			users[i] = *out
		}
	}
	return &usersListOutputDTO{
		Users:      users,
		TotalCount: p.TotalCount,
		TotalPages: p.TotalPages,
	}
}

func mapLinkTypeToOutput(t *domain.TrackerLinkType) *linkTypeOutputDTO {
	if t == nil {
		return nil
	}
	return &linkTypeOutputDTO{
		ID:      t.ID,
		Inward:  t.Inward,
		Outward: t.Outward,
	}
}

func mapLinkedIssueToOutput(i *domain.TrackerLinkedIssue) *linkedIssueOutputDTO {
	if i == nil {
		return nil
	}
	return &linkedIssueOutputDTO{
		Self:    i.Self,
		ID:      i.ID,
		Key:     i.Key,
		Display: i.Display,
	}
}

func mapLinkToOutput(l *domain.TrackerLink) linkOutputDTO {
	return linkOutputDTO{
		ID:        l.ID,
		Self:      l.Self,
		Type:      mapLinkTypeToOutput(l.Type),
		Direction: l.Direction,
		Object:    mapLinkedIssueToOutput(l.Object),
		CreatedBy: mapUserToOutput(l.CreatedBy),
		UpdatedBy: mapUserToOutput(l.UpdatedBy),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func mapLinksToOutput(links []domain.TrackerLink) *linksListOutputDTO {
	result := make([]linkOutputDTO, len(links))
	for i, link := range links {
		result[i] = mapLinkToOutput(&link)
	}
	return &linksListOutputDTO{Links: result}
}

func mapChangelogFieldToOutput(f domain.TrackerChangelogFieldChange) changelogFieldOutputDTO {
	return changelogFieldOutputDTO{
		Field: f.Field,
		From:  f.From,
		To:    f.To,
	}
}

func mapChangelogEntryToOutput(e *domain.TrackerChangelogEntry) changelogEntryOutputDTO {
	fields := make([]changelogFieldOutputDTO, len(e.Fields))
	for i, f := range e.Fields {
		fields[i] = mapChangelogFieldToOutput(f)
	}
	return changelogEntryOutputDTO{
		ID:        e.ID,
		Self:      e.Self,
		Issue:     mapLinkedIssueToOutput(e.Issue),
		UpdatedAt: e.UpdatedAt,
		UpdatedBy: mapUserToOutput(e.UpdatedBy),
		Type:      e.Type,
		Transport: e.Transport,
		Fields:    fields,
	}
}

func mapChangelogToOutput(entries []domain.TrackerChangelogEntry) *changelogOutputDTO {
	result := make([]changelogEntryOutputDTO, len(entries))
	for i, entry := range entries {
		result[i] = mapChangelogEntryToOutput(&entry)
	}
	return &changelogOutputDTO{Entries: result}
}

func mapProjectCommentToOutput(c *domain.TrackerProjectComment) projectCommentOutputDTO {
	return projectCommentOutputDTO{
		ID:        c.ID,
		LongID:    c.LongID,
		Self:      c.Self,
		Text:      c.Text,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		CreatedBy: mapUserToOutput(c.CreatedBy),
		UpdatedBy: mapUserToOutput(c.UpdatedBy),
	}
}

func mapProjectCommentsToOutput(comments []domain.TrackerProjectComment) *projectCommentsListOutputDTO {
	result := make([]projectCommentOutputDTO, len(comments))
	for i, comment := range comments {
		result[i] = mapProjectCommentToOutput(&comment)
	}
	return &projectCommentsListOutputDTO{Comments: result}
}
