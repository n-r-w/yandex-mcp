package tracker

import "github.com/n-r-w/yandex-mcp/internal/domain"

func issueToTrackerIssue(dto issueDTO) domain.TrackerIssue {
	return domain.TrackerIssue{
		Self:            dto.Self,
		ID:              dto.ID.String(),
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

func statusToTrackerStatus(dto *statusDTO) *domain.TrackerStatus {
	if dto == nil {
		return nil
	}
	return &domain.TrackerStatus{
		Self:    dto.Self,
		ID:      dto.ID.String(),
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func typeToTrackerIssueType(dto *typeDTO) *domain.TrackerIssueType {
	if dto == nil {
		return nil
	}
	return &domain.TrackerIssueType{
		Self:    dto.Self,
		ID:      dto.ID.String(),
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func prioToTrackerPriority(dto *prioDTO) *domain.TrackerPriority {
	if dto == nil {
		return nil
	}
	return &domain.TrackerPriority{
		Self:    dto.Self,
		ID:      dto.ID.String(),
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func queueToTrackerQueue(dto *queueDTO) *domain.TrackerQueue {
	if dto == nil {
		return nil
	}
	return &domain.TrackerQueue{
		Self:           dto.Self,
		ID:             dto.ID.String(),
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

func userToTrackerUser(dto *userDTO) *domain.TrackerUser {
	if dto == nil {
		return nil
	}
	return &domain.TrackerUser{
		Self:        dto.Self,
		ID:          dto.ID.String(),
		UID:         dto.UID.String(),
		Login:       dto.Login,
		Display:     dto.Display,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		Email:       dto.Email,
		CloudUID:    dto.CloudUID.String(),
		PassportUID: dto.PassportUID.String(),
	}
}

func transitionToTrackerTransition(dto transitionDTO) domain.TrackerTransition {
	return domain.TrackerTransition{
		ID:      dto.ID.String(),
		Display: dto.Display,
		Self:    dto.Self,
		To:      statusToTrackerStatus(dto.To),
	}
}

func commentToTrackerComment(dto commentDTO) domain.TrackerComment {
	return domain.TrackerComment{
		ID:        dto.ID.String(),
		LongID:    dto.LongID.String(),
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

func searchIssuesResultToTrackerIssuesPage(
	issues []issueDTO,
	totalCount int,
	totalPages int,
	scrollID string,
	scrollToken string,
	nextLink string,
) domain.TrackerIssuesPage {
	trackerIssues := make([]domain.TrackerIssue, len(issues))
	for i, issue := range issues {
		trackerIssues[i] = issueToTrackerIssue(issue)
	}
	return domain.TrackerIssuesPage{
		Issues:      trackerIssues,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		ScrollID:    scrollID,
		ScrollToken: scrollToken,
		NextLink:    nextLink,
	}
}

func listQueuesResultToTrackerQueuesPage(
	queues []queueDTO,
	totalCount int,
	totalPages int,
) domain.TrackerQueuesPage {
	trackerQueues := make([]domain.TrackerQueue, len(queues))
	for i, queue := range queues {
		trackerQueues[i] = *queueToTrackerQueue(&queue)
	}
	return domain.TrackerQueuesPage{
		Queues:     trackerQueues,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
}

func listCommentsResultToTrackerCommentsPage(
	comments []commentDTO,
	nextLink string,
) domain.TrackerCommentsPage {
	trackerComments := make([]domain.TrackerComment, len(comments))
	for i, comment := range comments {
		trackerComments[i] = commentToTrackerComment(comment)
	}
	return domain.TrackerCommentsPage{
		Comments: trackerComments,
		NextLink: nextLink,
	}
}

func attachmentToTrackerAttachment(dto attachmentDTO) domain.TrackerAttachment {
	var metadata *domain.TrackerAttachmentMetadata
	if dto.Metadata != nil {
		metadata = &domain.TrackerAttachmentMetadata{
			Size: dto.Metadata.Size,
		}
	}

	return domain.TrackerAttachment{
		ID:           dto.ID.String(),
		Name:         dto.Name,
		ContentURL:   dto.Content,
		ThumbnailURL: dto.Thumbnail,
		Mimetype:     dto.Mimetype,
		Size:         dto.Size,
		CreatedAt:    dto.CreatedAt,
		CreatedBy:    userToTrackerUser(dto.CreatedBy),
		Metadata:     metadata,
	}
}

func queueDetailToTrackerQueueDetail(dto queueDetailDTO) domain.TrackerQueueDetail {
	return domain.TrackerQueueDetail{
		Self:            dto.Self,
		ID:              dto.ID.String(),
		Key:             dto.Key,
		Display:         dto.Display,
		Name:            dto.Name,
		Description:     dto.Description,
		Version:         dto.Version,
		Lead:            userToTrackerUser(dto.Lead),
		AssignAuto:      dto.AssignAuto,
		AllowExternals:  dto.AllowExternals,
		DenyVoting:      dto.DenyVoting,
		DefaultType:     typeToTrackerIssueType(dto.DefaultType),
		DefaultPriority: prioToTrackerPriority(dto.DefaultPriority),
	}
}

func userDetailToTrackerUserDetail(dto userDetailDTO) domain.TrackerUserDetail {
	return domain.TrackerUserDetail{
		Self:        dto.Self,
		ID:          dto.ID.String(),
		UID:         dto.UID.String(),
		TrackerUID:  dto.TrackerUID.String(),
		Login:       dto.Login,
		Display:     dto.Display,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		Email:       dto.Email,
		CloudUID:    dto.CloudUID.String(),
		PassportUID: dto.PassportUID.String(),
		HasLicense:  dto.HasLicense,
		Dismissed:   dto.Dismissed,
		External:    dto.External,
	}
}

func linkToTrackerLink(dto linkDTO) domain.TrackerLink {
	return domain.TrackerLink{
		ID:        dto.ID.String(),
		Self:      dto.Self,
		Type:      linkTypeToTrackerLinkType(dto.Type),
		Direction: dto.Direction,
		Object:    linkedIssueToTrackerLinkedIssue(dto.Object),
		CreatedBy: userToTrackerUser(dto.CreatedBy),
		UpdatedBy: userToTrackerUser(dto.UpdatedBy),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

func linkTypeToTrackerLinkType(dto *linkTypeDTO) *domain.TrackerLinkType {
	if dto == nil {
		return nil
	}
	return &domain.TrackerLinkType{
		ID:      dto.ID.String(),
		Inward:  dto.Inward,
		Outward: dto.Outward,
	}
}

func linkedIssueToTrackerLinkedIssue(dto *linkedIssueDTO) *domain.TrackerLinkedIssue {
	if dto == nil {
		return nil
	}
	return &domain.TrackerLinkedIssue{
		Self:    dto.Self,
		ID:      dto.ID.String(),
		Key:     dto.Key,
		Display: dto.Display,
	}
}

func changelogEntryToTrackerChangelogEntry(dto changelogEntryDTO) domain.TrackerChangelogEntry {
	fields := make([]domain.TrackerChangelogFieldChange, len(dto.Fields))
	for i, f := range dto.Fields {
		var fieldID string
		if f.Field != nil {
			fieldID = f.Field.ID.String()
		}
		fields[i] = domain.TrackerChangelogFieldChange{
			Field: fieldID,
			From:  f.From,
			To:    f.To,
		}
	}
	return domain.TrackerChangelogEntry{
		ID:        dto.ID.String(),
		Self:      dto.Self,
		Issue:     linkedIssueToTrackerLinkedIssue(dto.Issue),
		UpdatedAt: dto.UpdatedAt,
		UpdatedBy: userToTrackerUser(dto.UpdatedBy),
		Type:      dto.Type,
		Transport: dto.Transport,
		Fields:    fields,
	}
}

func projectCommentToTrackerProjectComment(dto projectCommentDTO) domain.TrackerProjectComment {
	return domain.TrackerProjectComment{
		ID:        dto.ID.String(),
		LongID:    dto.LongID.String(),
		Self:      dto.Self,
		Text:      dto.Text,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		CreatedBy: userToTrackerUser(dto.CreatedBy),
		UpdatedBy: userToTrackerUser(dto.UpdatedBy),
	}
}
