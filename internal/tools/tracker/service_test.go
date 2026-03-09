package tracker

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/domain"
	mcpserver "github.com/n-r-w/yandex-mcp/internal/server"
)

// TestRegistrator_Register_RegistersOnlyEnabledTrackerTools verifies the real tracker registrator wiring.
func TestRegistrator_Register_RegistersOnlyEnabledTrackerTools(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockAdapter := NewMockITrackerAdapter(ctrl)
	registrator := NewRegistrator(
		mockAdapter,
		[]domain.TrackerTool{
			domain.TrackerToolIssueGet,
			domain.TrackerToolBoardsList,
			domain.TrackerToolAttachmentGet,
		},
		defaultAttachExtensions,
		defaultAttachViewExts,
		defaultAttachDirs,
	)

	srv, err := mcpserver.New("v1.0.0", []mcpserver.IToolsRegistrator{registrator})
	require.NoError(t, err)

	client := mcp.NewClient(
		&mcp.Implementation{ //nolint:exhaustruct // test helper uses only required metadata
			Name:    "test-client",
			Version: "v1.0.0",
		},
		nil,
	)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	_, err = srv.Connect(t.Context(), serverTransport)
	require.NoError(t, err)

	session, err := client.Connect(t.Context(), clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	toolNames := make([]string, 0, 3)
	boardsToolHasSchema := false
	for tool, toolErr := range session.Tools(t.Context(), nil) {
		require.NoError(t, toolErr)
		toolNames = append(toolNames, tool.Name)
		if tool.Name == domain.TrackerToolBoardsList.String() {
			boardsToolHasSchema = tool.InputSchema != nil
		}
	}

	assert.Len(t, toolNames, 3)
	assert.Contains(t, toolNames, domain.TrackerToolIssueGet.String())
	assert.Contains(t, toolNames, domain.TrackerToolBoardsList.String())
	assert.Contains(t, toolNames, domain.TrackerToolAttachmentGet.String())
	assert.NotContains(t, toolNames, domain.TrackerToolQueueGet.String())
	assert.True(t, boardsToolHasSchema)
}
