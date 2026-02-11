package itest

import (
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

var (
	defaultAttachExtensions = []string{"txt"}
	defaultAttachViewExts   = []string{"txt"}
	defaultAttachDirs       []string
)

func listToolNames(t *testing.T, srv *server.Server) []string {
	t.Helper()

	ctx := t.Context()

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
		wikitools.NewRegistrator(wikiMock, domain.WikiAllTools()),
		trackertools.NewRegistrator(
			trackerMock,
			domain.TrackerAllTools(),
			defaultAttachExtensions,
			defaultAttachViewExts,
			defaultAttachDirs,
		),
	}

	srv, err := server.New("v1.0.0", registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)

	expectedTools := make([]string, 0, len(domain.WikiAllTools())+len(domain.TrackerAllTools()))
	for _, tool := range domain.WikiAllTools() {
		expectedTools = append(expectedTools, tool.String())
	}
	for _, tool := range domain.TrackerAllTools() {
		expectedTools = append(expectedTools, tool.String())
	}

	assert.ElementsMatch(t, expectedTools, toolNames)
}

func TestServerIntegration_AllowlistGating_ReducedList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	wikiMock := wikitools.NewMockIWikiAdapter(ctrl)
	trackerMock := trackertools.NewMockITrackerAdapter(ctrl)

	wikiTools := []domain.WikiTool{domain.WikiToolPageGetBySlug}
	trackerTools := []domain.TrackerTool{domain.TrackerToolIssueGet}

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiMock, wikiTools),
		trackertools.NewRegistrator(
			trackerMock,
			trackerTools,
			defaultAttachExtensions,
			defaultAttachViewExts,
			defaultAttachDirs,
		),
	}

	srv, err := server.New("v1.0.0", registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)

	assert.ElementsMatch(t, []string{
		domain.WikiToolPageGetBySlug.String(),
		domain.TrackerToolIssueGet.String(),
	}, toolNames)

	assert.NotContains(t, toolNames, domain.WikiToolPageGetByID.String())
	assert.NotContains(t, toolNames, domain.TrackerToolIssueSearch.String())
}

func TestServerIntegration_EmptyAllowlist_NoToolsRegistered(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	wikiMock := wikitools.NewMockIWikiAdapter(ctrl)
	trackerMock := trackertools.NewMockITrackerAdapter(ctrl)

	registrators := []server.IToolsRegistrator{
		wikitools.NewRegistrator(wikiMock, nil),
		trackertools.NewRegistrator(
			trackerMock,
			nil,
			defaultAttachExtensions,
			defaultAttachViewExts,
			defaultAttachDirs,
		),
	}

	srv, err := server.New("v1.0.0", registrators)
	require.NoError(t, err)

	toolNames := listToolNames(t, srv)
	assert.Empty(t, toolNames)
}
