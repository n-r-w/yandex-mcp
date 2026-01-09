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
}
