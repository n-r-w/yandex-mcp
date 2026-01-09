package tracker

import "github.com/n-r-w/yandex-mcp/internal/domain"

func issueToTrackerIssue(dto Issue) domain.TrackerIssue {
	return domain.TrackerIssue{
		Self:            dto.Self,
		ID:              dto.ID,
		Key:             dto.Key,
		Version:         dto.Version,
		Summary:         dto.Summary,
		Description:     dto.Description,
		StatusStartTime: dto.StatusStartTime,
		CreatedAt:       dto.CreatedAt,
		UpdatedAt:       dto.UpdatedAt,
		ResolvedAt:      dto.ResolvedAt,
		Status:          statusToTrackerStatus(dto.Status),
		Type:            typeToTrackerIssueType(dto.Type),
		Priority:        prioToTrackerPriority(dto.Priority),
		Queue:           queueToTrackerQueue(dto.Queue),
		Assignee:        userToTrackerUser(dto.Assignee),
		CreatedBy:       userToTrackerUser(dto.CreatedBy),
		UpdatedBy:       userToTrackerUser(dto.UpdatedBy),
		Votes:           dto.Votes,
		Favorite:        dto.Favorite,
	}
}

func statusToTrackerStatus(dto *Status) *domain.TrackerStatus {
	if dto == nil {
		return nil
	}
	return &domain.TrackerStatus{
		Self:    dto.Self,
		ID:      dto.ID,
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func typeToTrackerIssueType(dto *Type) *domain.TrackerIssueType {
	if dto == nil {
		return nil
	}
	return &domain.TrackerIssueType{
		Self:    dto.Self,
		ID:      dto.ID,
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func prioToTrackerPriority(dto *Prio) *domain.TrackerPriority {
	if dto == nil {
		return nil
	}
	return &domain.TrackerPriority{
		Self:    dto.Self,
		ID:      dto.ID,
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func queueToTrackerQueue(dto *Queue) *domain.TrackerQueue {
	if dto == nil {
		return nil
	}
	return &domain.TrackerQueue{
		Self:           dto.Self,
		ID:             dto.ID,
		Key:            dto.Key,
		Display:        dto.Display,
		Name:           dto.Name,
		Version:        dto.Version,
		Lead:           userToTrackerUser(dto.Lead),
		AssignAuto:     dto.AssignAuto,
		AllowExternals: dto.AllowExternals,
		DenyVoting:     dto.DenyVoting,
	}
}

func userToTrackerUser(dto *User) *domain.TrackerUser {
	if dto == nil {
		return nil
	}
	return &domain.TrackerUser{
		Self:        dto.Self,
		ID:          dto.ID,
		UID:         dto.UID,
		Login:       dto.Login,
		Display:     dto.Display,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		Email:       dto.Email,
		CloudUID:    dto.CloudUID,
		PassportUID: dto.PassportUID,
	}
}

func transitionToTrackerTransition(dto Transition) domain.TrackerTransition {
	return domain.TrackerTransition{
		ID:      dto.ID,
		Display: dto.Display,
		Self:    dto.Self,
		To:      statusToTrackerStatus(dto.To),
	}
}

func commentToTrackerComment(dto Comment) domain.TrackerComment {
	return domain.TrackerComment{
		ID:        dto.ID,
		LongID:    dto.LongID,
		Self:      dto.Self,
		Text:      dto.Text,
		Version:   dto.Version,
		Type:      dto.Type,
		Transport: dto.Transport,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		CreatedBy: userToTrackerUser(dto.CreatedBy),
		UpdatedBy: userToTrackerUser(dto.UpdatedBy),
	}
}

func searchIssuesResultToTrackerIssuesPage(dto SearchIssuesResult) domain.TrackerIssuesPage {
	issues := make([]domain.TrackerIssue, len(dto.Issues))
	for i, issue := range dto.Issues {
		issues[i] = issueToTrackerIssue(issue)
	}
	return domain.TrackerIssuesPage{
		Issues:      issues,
		TotalCount:  dto.TotalCount,
		TotalPages:  dto.TotalPages,
		ScrollID:    dto.ScrollID,
		ScrollToken: dto.ScrollToken,
		NextLink:    dto.NextLink,
	}
}

func listQueuesResultToTrackerQueuesPage(dto ListQueuesResult) domain.TrackerQueuesPage {
	queues := make([]domain.TrackerQueue, len(dto.Queues))
	for i, queue := range dto.Queues {
		queues[i] = *queueToTrackerQueue(&queue)
	}
	return domain.TrackerQueuesPage{
		Queues:     queues,
		TotalCount: dto.TotalCount,
		TotalPages: dto.TotalPages,
	}
}

func listCommentsResultToTrackerCommentsPage(dto ListCommentsResult) domain.TrackerCommentsPage {
	comments := make([]domain.TrackerComment, len(dto.Comments))
	for i, comment := range dto.Comments {
		comments[i] = commentToTrackerComment(comment)
	}
	return domain.TrackerCommentsPage{
		Comments: comments,
		NextLink: dto.NextLink,
	}
}
