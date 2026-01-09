// Package tracker provides MCP tool handlers for Yandex Tracker operations.
package tracker

import (
	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// Mapping functions from domain models to tool outputs.

func mapIssueToOutput(i *domain.TrackerIssue) *IssueOutput {
	if i == nil {
		return nil
	}

	return &IssueOutput{
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

func mapStatusToOutput(s *domain.TrackerStatus) *StatusOutput {
	if s == nil {
		return nil
	}
	return &StatusOutput{
		Self:    s.Self,
		ID:      s.ID,
		Key:     s.Key,
		Display: s.Display,
	}
}

func mapTypeToOutput(t *domain.TrackerIssueType) *TypeOutput {
	if t == nil {
		return nil
	}
	return &TypeOutput{
		Self:    t.Self,
		ID:      t.ID,
		Key:     t.Key,
		Display: t.Display,
	}
}

func mapPriorityToOutput(p *domain.TrackerPriority) *PriorityOutput {
	if p == nil {
		return nil
	}
	return &PriorityOutput{
		Self:    p.Self,
		ID:      p.ID,
		Key:     p.Key,
		Display: p.Display,
	}
}

func mapQueueToOutput(q *domain.TrackerQueue) *QueueOutput {
	if q == nil {
		return nil
	}
	return &QueueOutput{
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

func mapUserToOutput(u *domain.TrackerUser) *UserOutput {
	if u == nil {
		return nil
	}
	return &UserOutput{
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

func mapTransitionToOutput(t *domain.TrackerTransition) *TransitionOutput {
	if t == nil {
		return nil
	}
	return &TransitionOutput{
		ID:      t.ID,
		Display: t.Display,
		Self:    t.Self,
		To:      mapStatusToOutput(t.To),
	}
}

func mapCommentToOutput(c *domain.TrackerComment) *CommentOutput {
	if c == nil {
		return nil
	}
	return &CommentOutput{
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

func mapSearchResultToOutput(r *domain.TrackerIssuesPage) *SearchIssuesOutput {
	if r == nil {
		return nil
	}

	issues := make([]IssueOutput, len(r.Issues))
	for i, issue := range r.Issues {
		out := mapIssueToOutput(&issue)
		if out != nil {
			issues[i] = *out
		}
	}

	return &SearchIssuesOutput{
		Issues:      issues,
		TotalCount:  r.TotalCount,
		TotalPages:  r.TotalPages,
		ScrollID:    r.ScrollID,
		ScrollToken: r.ScrollToken,
		NextLink:    r.NextLink,
	}
}

func mapTransitionsToOutput(transitions []domain.TrackerTransition) *TransitionsListOutput {
	out := make([]TransitionOutput, len(transitions))
	for i, t := range transitions {
		mapped := mapTransitionToOutput(&t)
		if mapped != nil {
			out[i] = *mapped
		}
	}
	return &TransitionsListOutput{Transitions: out}
}

func mapQueuesResultToOutput(r *domain.TrackerQueuesPage) *QueuesListOutput {
	if r == nil {
		return nil
	}

	queues := make([]QueueOutput, len(r.Queues))
	for i, q := range r.Queues {
		out := mapQueueToOutput(&q)
		if out != nil {
			queues[i] = *out
		}
	}

	return &QueuesListOutput{
		Queues:     queues,
		TotalCount: r.TotalCount,
		TotalPages: r.TotalPages,
	}
}

func mapCommentsResultToOutput(r *domain.TrackerCommentsPage) *CommentsListOutput {
	if r == nil {
		return nil
	}

	comments := make([]CommentOutput, len(r.Comments))
	for i, c := range r.Comments {
		out := mapCommentToOutput(&c)
		if out != nil {
			comments[i] = *out
		}
	}

	return &CommentsListOutput{
		Comments: comments,
		NextLink: r.NextLink,
	}
}
