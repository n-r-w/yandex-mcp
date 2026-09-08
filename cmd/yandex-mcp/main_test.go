package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// The native test executable also runs the real application entry point in a child.
func TestMain(m *testing.M) {
	if os.Getenv("YANDEX_MCP_TEST_PROCESS") == "1" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
		if err := run("test", nil); err != nil {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// A separate MCP process has no yc in PATH. Its real stdio request reaches an
// HTTP auth-agent and returns its complete diagnostic in both the tool result
// and stderr log, including messages larger than 4096 bytes. Only yc is mocked.
func TestRemoteMCPPreservesErrorAndLogs(t *testing.T) {
	t.Setenv("YANDEX_MCP_TEST_PROCESS", "1")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "remote")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "test-org")
	t.Setenv("YANDEX_CLI_PROFILE", "work")
	t.Setenv("PATH", "")
	diagnostic := "original yc error " + strings.Repeat("diagnostic-", 600)
	source := authagent.NewMockITokenSource(gomock.NewController(t))
	source.EXPECT().Acquire(gomock.Any(), "work").Return("", errors.New(diagnostic))
	service := authagent.New(source)
	defer service.Close()
	endpoint := httptest.NewServer(service)
	defer endpoint.Close()
	t.Setenv("YANDEX_MCP_AUTH_AGENT_PORT", strconv.Itoa(endpoint.Listener.Addr().(*net.TCPAddr).Port))
	executable, err := os.Executable()
	require.NoError(t, err)
	command := exec.CommandContext(t.Context(), executable)
	var logs bytes.Buffer
	command.Stderr = &logs
	//nolint:exhaustruct_v5 // test client identity
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	//nolint:exhaustruct_v5 // default process shutdown duration
	session, err := client.Connect(t.Context(), &mcp.CommandTransport{Command: command}, nil)
	require.NoError(t, err)
	defer func() { _ = session.Close() }()
	//nolint:exhaustruct_v5 // no request metadata
	result, err := session.CallTool(
		t.Context(),
		&mcp.CallToolParams{Name: "wiki_page_get_by_id", Arguments: map[string]any{"page_id": "1"}},
	)
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, text.Text, diagnostic)
	require.NoError(t, session.Close())
	require.Contains(t, logs.String(), diagnostic)
}
