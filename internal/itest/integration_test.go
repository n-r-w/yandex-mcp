package itest

import (
	"context"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	"github.com/n-r-w/yandex-mcp/internal/server"
	trackertools "github.com/n-r-w/yandex-mcp/internal/tools/tracker"
	wikitools "github.com/n-r-w/yandex-mcp/internal/tools/wiki"
)

func listToolNames(t *testing.T, srv *server.Server) []string {
	t.Helper()

	ctx := context.Background()

	client := mcp.NewClient(
		&mcp.Implementation{ //nolint:exhaustruct // optional fields use defaults
			Name:    "test-client",
			Version: "1.0.0",
		},
		nil,
	)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	_, err := srv.Connect(ctx, serverTransport)
	require.NoError(t, err)

	session, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()

	toolNames := make([]string, 0)
	for tool, err := range session.Tools(ctx, nil) {
		require.NoError(t, err)
		toolNames = append(toolNames, tool.Name)
	}

	return toolNames
}

func TestServerIntegration_ReadOnlyToolsRegistered(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	wikiMock := wikitools.NewMockIWikiAdapter(ctrl)
	trackerMock := trackertools.NewMockITrackerAdapter(ctrl)

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiMock, domain.WikiReadOnlyTools()),
		trackertools.NewRegistrator(trackerMock, domain.TrackerReadOnlyTools()),
	}

	srv, err := server.New(registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)

	expectedTools := make([]string, 0, len(domain.WikiReadOnlyTools())+len(domain.TrackerReadOnlyTools()))
	for _, tool := range domain.WikiReadOnlyTools() {
		expectedTools = append(expectedTools, tool.String())
	}
	for _, tool := range domain.TrackerReadOnlyTools() {
		expectedTools = append(expectedTools, tool.String())
	}

	assert.Len(t, toolNames, len(expectedTools), "should have exactly %d tools", len(expectedTools))
	for _, expected := range expectedTools {
		assert.True(t, slices.Contains(toolNames, expected), "tool %q should be registered", expected)
	}
}

func TestServerIntegration_AllowlistGating_ReducedList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	wikiMock := wikitools.NewMockIWikiAdapter(ctrl)
	trackerMock := trackertools.NewMockITrackerAdapter(ctrl)

	// Register only a subset of tools.
	wikiTools := []domain.WikiTool{domain.WikiToolPageGetBySlug}
	trackerTools := []domain.TrackerTool{domain.TrackerToolIssueGet}

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiMock, wikiTools),
		trackertools.NewRegistrator(trackerMock, trackerTools),
	}

	srv, err := server.New(registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)

	// Only the allowed tools should be registered.
	assert.ElementsMatch(t, []string{
		domain.WikiToolPageGetBySlug.String(),
		domain.TrackerToolIssueGet.String(),
	}, toolNames)

	// Verify excluded tools are not present.
	assert.False(t, slices.Contains(toolNames, domain.WikiToolPageGetByID.String()))
	assert.False(t, slices.Contains(toolNames, domain.TrackerToolIssueSearch.String()))
}

func TestServerIntegration_EmptyAllowlist_NoToolsRegistered(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	wikiMock := wikitools.NewMockIWikiAdapter(ctrl)
	trackerMock := trackertools.NewMockITrackerAdapter(ctrl)

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiMock, nil),
		trackertools.NewRegistrator(trackerMock, nil),
	}

	srv, err := server.New(registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)
	assert.Empty(t, toolNames)
}
