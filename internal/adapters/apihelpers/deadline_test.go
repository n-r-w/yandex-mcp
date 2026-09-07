package apihelpers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/server"
)

// A real MCP call shares one deadline across initial token acquisition, a 401,
// forced refresh and the API retry. Mocks model each stage; fake time makes the
// five-second budget deterministic. The second API request must get two seconds.
func TestMCPDeadlineIncludesTokenAndRetry(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		provider := NewMockITokenProvider(ctrl)
		doer := NewMockIHTTPDoer(ctrl)
		client := newTestAPIClient(doer, provider)
		start := time.Now()
		deadline := start.Add(5 * time.Second)
		for _, force := range []bool{false, true} {
			provider.EXPECT().Token(gomock.Any(), force).DoAndReturn(func(ctx context.Context, _ bool) (string, error) {
				actual, ok := ctx.Deadline()
				assert.True(t, ok)
				assert.Equal(t, deadline, actual)
				time.Sleep(time.Second)
				return "token", nil
			})
		}
		gomock.InOrder(
			doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(request *http.Request) (*http.Response, error) {
				actual, _ := request.Context().Deadline()
				assert.Equal(t, deadline, actual)
				time.Sleep(time.Second)
				//nolint:exhaustruct_v5 // only response body and status are used
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(strings.NewReader("")),
				}, nil
			}),
			doer.EXPECT().Do(gomock.Any()).DoAndReturn(func(request *http.Request) (*http.Response, error) {
				actual, _ := request.Context().Deadline()
				assert.Equal(t, deadline, actual)
				assert.Equal(t, 2*time.Second, time.Until(actual))
				<-request.Context().Done()
				return nil, request.Context().Err()
			}),
		)
		finished := make(chan struct{})
		registrar := server.NewMockIToolsRegistrator(ctrl)
		registrar.EXPECT().Register(gomock.Any()).DoAndReturn(func(srv *mcp.Server) error {
			//nolint:exhaustruct_v5 // schema inferred from empty input
			mcp.AddTool(
				srv,
				&mcp.Tool{Name: "deadline"},
				func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
					defer close(finished)
					_, _, err := client.DoGETRaw(ctx, "/test", "test")
					return nil, struct{}{}, err
				},
			)
			return nil
		})
		srv, err := server.New("test", []server.IToolsRegistrator{registrar}, 5*time.Second)
		require.NoError(t, err)
		serverTransport, clientTransport := mcp.NewInMemoryTransports()
		serverSession, err := srv.Connect(t.Context(), serverTransport)
		require.NoError(t, err)
		defer func() { _ = serverSession.Close() }()
		//nolint:exhaustruct_v5 // test client identity
		mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
		session, err := mcpClient.Connect(t.Context(), clientTransport, nil)
		require.NoError(t, err)
		defer func() { _ = session.Close() }()
		//nolint:exhaustruct_v5 // no tool arguments
		result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "deadline"})
		require.NoError(t, err)
		require.True(t, result.IsError)
		require.Equal(t, 5*time.Second, time.Since(start))
		<-finished
	})
}
