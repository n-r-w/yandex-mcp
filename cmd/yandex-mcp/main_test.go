package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/config"
	"github.com/n-r-w/yandex-mcp/internal/server/authagent"
)

// The native test executable also runs the real application entry point in a child,
// and acts as yc when a test copies it under that name.
func TestMain(m *testing.M) {
	if os.Getenv("YANDEX_MCP_TEST_YC") == "1" && len(os.Args) >= 3 && os.Args[1] == "iam" {
		profile := "active"
		if len(os.Args) == 5 && os.Args[3] == "--profile" {
			profile = os.Args[4]
		}
		_, _ = fmt.Fprintln(os.Stdout, testIAMToken(profile))
		os.Exit(0)
	}
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

// testIAMToken is the IAM token the fake yc prints for a profile.
func testIAMToken(profile string) string {
	return "t1." + profile + "." + strings.Repeat("a", 86)
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
	ctrl := gomock.NewController(t)
	source := authagent.NewMockITokenSource(ctrl)
	source.EXPECT().Acquire(gomock.Any(), "work").Return("", errors.New(diagnostic))
	service := authagent.New(source, authagent.NewMockITokenSource(ctrl))
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

// A separate MCP process with the oauth token type has no yc in PATH and no auth-agent.
// Its real stdio request must read the configured token file. The file is absent, so the tool
// result names that exact path and the stderr log carries the file source diagnostic.
// Without the token type selection the process would fail on the missing yc instead.
func TestOAuthMCPReadsConfiguredTokenFile(t *testing.T) {
	const diagnostic = "read OAuth token file"
	tokenFile := filepath.Join(t.TempDir(), "absent-oauth-token")
	t.Setenv("YANDEX_MCP_TEST_PROCESS", "1")
	t.Setenv("YANDEX_MCP_TOKEN_SOURCE", "local")
	t.Setenv("YANDEX_MCP_TOKEN_TYPE", "oauth")
	t.Setenv("YANDEX_OAUTH_TOKEN_FILE", tokenFile)
	t.Setenv("YANDEX_ORG_ID", "1234567")
	t.Setenv("YANDEX_CLOUD_ORG_ID", "")
	t.Setenv("PATH", "")
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
	require.Contains(t, text.Text, tokenFile)
	require.NoError(t, session.Close())
	require.Contains(t, logs.String(), diagnostic)
}

// newTokenProvider composes the acquisition location with the token type. Every combination
// delivers the token of its own source: fake yc output for IAM and file content for OAuth,
// obtained on this machine for local acquisition and through a real auth-agent handler for
// remote acquisition. IAM profile "work" and OAuth profile "work" resolve to different tokens,
// OAuth profiles select different files, and a failing source is never replaced by another one.
// The auth-agent is built by newAuthAgent, the composition that the auth-agent command of run uses.
func TestTokenProviderCombinations(t *testing.T) {
	executable, err := os.Executable()
	require.NoError(t, err)
	binary, err := os.ReadFile(executable)
	require.NoError(t, err)
	name := "yc"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	ycPath := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(ycPath, binary, 0o700))
	t.Setenv("PATH", filepath.Dir(ycPath))
	t.Setenv("YANDEX_MCP_TEST_YC", "1")

	localFile := filepath.Join(t.TempDir(), "local-token")
	require.NoError(t, os.WriteFile(localFile, []byte("local-oauth-token"), 0o600))
	absentFile := filepath.Join(t.TempDir(), "absent-token")
	agentDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(agentDir, "work"), []byte("agent-work-token"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(agentDir, "personal"), []byte("agent-personal-token"), 0o600))

	agent := newAuthAgent(&authagent.Config{Port: 0, YCPath: ycPath, SSHTarget: "", OAuthDir: agentDir})
	defer agent.Close()
	endpoint := httptest.NewServer(agent)
	defer endpoint.Close()
	port := endpoint.Listener.Addr().(*net.TCPAddr).Port

	const local, remote = config.TokenSourceLocal, config.TokenSourceRemote
	const iam, oauth = config.TokenTypeIAM, config.TokenTypeOAuth
	tests := []struct {
		name, source, tokenType        string
		cliProfile, oauthProfile, file string
		want, wantErr                  string
	}{
		{"local iam", local, iam, "local-profile", "", "", testIAMToken("local-profile"), ""},
		{"remote iam", remote, iam, "work", "", "", testIAMToken("work"), ""},
		{"local oauth", local, oauth, "work", "", localFile, "local-oauth-token", ""},
		{"remote oauth", remote, oauth, "work", "work", "", "agent-work-token", ""},
		{"remote oauth other profile", remote, oauth, "work", "personal", "", "agent-personal-token", ""},
		{"local oauth failure keeps yc unused", local, oauth, "work", "", absentFile, "", "read OAuth token file"},
		{"remote oauth failure keeps yc unused", remote, oauth, "work", "absent", "", "", "unknown OAuth profile"},
	}
	for _, tt := range tests {
		//nolint:exhaustruct_v5 // only token acquisition settings matter
		cfg := &config.Config{
			TokenSource:           tt.source,
			TokenType:             tt.tokenType,
			CLIProfile:            tt.cliProfile,
			OAuthProfile:          tt.oauthProfile,
			OAuthTokenFile:        tt.file,
			AuthAgentPort:         port,
			IAMTokenRefreshPeriod: time.Hour,
		}
		provider, closeProvider, providerErr := newTokenProvider(cfg)
		require.NoError(t, providerErr, tt.name)
		token, tokenErr := provider.Token(t.Context(), false)
		closeProvider()
		if tt.wantErr != "" {
			require.ErrorContains(t, tokenErr, tt.wantErr, tt.name)
			require.Empty(t, token, tt.name)
			continue
		}
		require.NoError(t, tokenErr, tt.name)
		require.Equal(t, tt.want, token, tt.name)
	}
}
