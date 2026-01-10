// Package tracker provides MCP tool handlers for Yandex Tracker operations.
package tracker

import (
	"context"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=interfaces.go -destination=mock_interfaces.go -package=tracker

// ITrackerAdapter defines the interface for Tracker adapter operations consumed by tools.
type ITrackerAdapter interface {
	GetIssue(ctx context.Context, issueID string, opts domain.TrackerGetIssueOpts) (*domain.TrackerIssue, error)
	SearchIssues(ctx context.Context, opts domain.TrackerSearchIssuesOpts) (*domain.TrackerIssuesPage, error)
	CountIssues(ctx context.Context, opts domain.TrackerCountIssuesOpts) (int, error)
	ListIssueTransitions(ctx context.Context, issueID string) ([]domain.TrackerTransition, error)
	ListQueues(ctx context.Context, opts domain.TrackerListQueuesOpts) (*domain.TrackerQueuesPage, error)
	ListIssueComments(
		ctx context.Context,
		issueID string,
		opts domain.TrackerListCommentsOpts,
	) (*domain.TrackerCommentsPage, error)
	CreateIssue(ctx context.Context, req *domain.TrackerIssueCreateRequest) (*domain.TrackerIssueCreateResponse, error)
	UpdateIssue(ctx context.Context, req *domain.TrackerIssueUpdateRequest) (*domain.TrackerIssueUpdateResponse, error)
	ExecuteTransition(
		ctx context.Context, req *domain.TrackerTransitionExecuteRequest,
	) (*domain.TrackerTransitionExecuteResponse, error)
	AddComment(ctx context.Context, req *domain.TrackerCommentAddRequest) (*domain.TrackerCommentAddResponse, error)
	UpdateComment(
		ctx context.Context, req *domain.TrackerCommentUpdateRequest,
	) (*domain.TrackerCommentUpdateResponse, error)
	DeleteComment(ctx context.Context, req *domain.TrackerCommentDeleteRequest) error
	ListIssueAttachments(ctx context.Context, issueID string) ([]domain.TrackerAttachment, error)
	DeleteAttachment(ctx context.Context, req *domain.TrackerAttachmentDeleteRequest) error
	GetQueue(
		ctx context.Context, queueID string, opts domain.TrackerGetQueueOpts,
	) (*domain.TrackerQueueDetail, error)
	CreateQueue(ctx context.Context, req *domain.TrackerQueueCreateRequest) (*domain.TrackerQueueCreateResponse, error)
	DeleteQueue(ctx context.Context, req *domain.TrackerQueueDeleteRequest) error
	RestoreQueue(
		ctx context.Context, req *domain.TrackerQueueRestoreRequest,
	) (*domain.TrackerQueueRestoreResponse, error)
	GetCurrentUser(ctx context.Context) (*domain.TrackerUserDetail, error)
	ListUsers(ctx context.Context, opts domain.TrackerListUsersOpts) (*domain.TrackerUsersPage, error)
	GetUser(ctx context.Context, userID string) (*domain.TrackerUserDetail, error)
	ListIssueLinks(ctx context.Context, issueID string) ([]domain.TrackerLink, error)
	CreateLink(ctx context.Context, req *domain.TrackerLinkCreateRequest) (*domain.TrackerLinkCreateResponse, error)
	DeleteLink(ctx context.Context, req *domain.TrackerLinkDeleteRequest) error
	GetIssueChangelog(
		ctx context.Context, issueID string, opts domain.TrackerGetChangelogOpts,
	) ([]domain.TrackerChangelogEntry, error)
	MoveIssue(ctx context.Context, req *domain.TrackerIssueMoveRequest) (*domain.TrackerIssueMoveResponse, error)
	ListProjectComments(
		ctx context.Context, projectID string, opts domain.TrackerListProjectCommentsOpts,
	) ([]domain.TrackerProjectComment, error)
}
