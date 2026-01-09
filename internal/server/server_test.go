package server

import (
	"context"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testWikiTools = []struct {
		name        string
		description string
	}{
		{domain.WikiToolPageGetBySlug.String(), "Retrieves a Yandex Wiki page by its slug"},
		{domain.WikiToolPageGetByID.String(), "Retrieves a Yandex Wiki page by its ID"},
		{domain.WikiToolResourcesList.String(), "Lists resources for a Yandex Wiki page"},
		{domain.WikiToolGridsList.String(), "Lists grids for a Yandex Wiki page"},
		{domain.WikiToolGridGet.String(), "Retrieves a Yandex Wiki grid"},
	}

	testTrackerTools = []struct {
		name        string
		description string
	}{
		{domain.TrackerToolIssueGet.String(), "Retrieves a Yandex Tracker issue"},
		{domain.TrackerToolIssueSearch.String(), "Searches Yandex Tracker issues"},
		{domain.TrackerToolIssueCount.String(), "Counts Yandex Tracker issues"},
		{domain.TrackerToolTransitionsList.String(), "Lists issue transitions"},
		{domain.TrackerToolQueuesList.String(), "Lists Yandex Tracker queues"},
		{domain.TrackerToolCommentsList.String(), "Lists issue comments"},
	}
)

func newWikiStubRegistrator(ctrl *gomock.Controller) IToolsRegistrator {
	mock := NewMockIToolsRegistrator(ctrl)
	mock.EXPECT().Register(gomock.Any()).DoAndReturn(func(srv *mcp.Server) error {
		for _, tool := range testWikiTools {
			mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
				Name:        tool.name,
				Description: tool.description,
			}, func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
				return nil, map[string]any{"status": "ok"}, nil
			})
		}
		return nil
	})
	return mock
}

func newTrackerStubRegistrator(ctrl *gomock.Controller) IToolsRegistrator {
	mock := NewMockIToolsRegistrator(ctrl)
	mock.EXPECT().Register(gomock.Any()).DoAndReturn(func(srv *mcp.Server) error {
		for _, tool := range testTrackerTools {
			mcp.AddTool(srv, &mcp.Tool{ //nolint:exhaustruct // optional fields use defaults
				Name:        tool.name,
				Description: tool.description,
			}, func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
				return nil, map[string]any{"status": "ok"}, nil
			})
		}
		return nil
	})
	return mock
}

func TestServer_ToolsRegistered(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	registrators := []IToolsRegistrator{
		newWikiStubRegistrator(ctrl),
		newTrackerStubRegistrator(ctrl),
	}

	srv, err := New(registrators)
	require.NoError(t, err)

	ctx := context.Background()

	// Connect to server using in-memory transport.
	client := mcp.NewClient(
		&mcp.Implementation{ //nolint:exhaustruct // optional fields use defaults
			Name:    "test-client",
			Version: "1.0.0",
		},
		nil,
	)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	_, err = srv.Connect(ctx, serverTransport)
	require.NoError(t, err)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	// List tools.
	toolNames := make([]string, 0, 11)
	for tool, err := range session.Tools(ctx, nil) {
		require.NoError(t, err)
		toolNames = append(toolNames, tool.Name)
	}

	// Verify all expected tools are registered.
	expectedTools := make([]string, 0, len(testWikiTools)+len(testTrackerTools))
	for _, tool := range testWikiTools {
		expectedTools = append(expectedTools, tool.name)
	}
	for _, tool := range testTrackerTools {
		expectedTools = append(expectedTools, tool.name)
	}

	assert.Len(t, toolNames, len(expectedTools), "should have exactly %d tools", len(expectedTools))
	for _, expected := range expectedTools {
		assert.True(t, slices.Contains(toolNames, expected), "tool %q should be registered", expected)
	}
}

func TestServerCreation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	srv, err := New([]IToolsRegistrator{newWikiStubRegistrator(ctrl)})
	require.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestServerCreation_EmptyRegistrators(t *testing.T) {
	t.Parallel()

	srv, err := New(nil)
	require.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestServerCreation_NoRegistrators(t *testing.T) {
	t.Parallel()

	srv, err := New([]IToolsRegistrator{})
	require.NoError(t, err)
	assert.NotNil(t, srv)
}

func TestServer_VersionInfo(t *testing.T) {
	t.Parallel()

	// Verify constants are properly set.
	assert.Equal(t, "yandex-mcp", serverName)
	assert.Equal(t, "1.0.0", serverVersion)
}

func TestServer_RegistrationError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockReg := NewMockIToolsRegistrator(ctrl)
	mockReg.EXPECT().Register(gomock.Any()).Return(assert.AnError)

	_, err := New([]IToolsRegistrator{mockReg})
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
