package wiki

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// DownloadAttachment returns the full attachment response body; the caller closes it.
func (c *Client) DownloadAttachment(ctx context.Context, attachmentPath string) (io.ReadCloser, error) {
	query := url.Values{"url": {attachmentPath}, "download": {"true"}}
	status, stream, err := c.apiClient.DoGETDownload(ctx, "/v1/pages/attachments/download_by_url?"+query.Encode())
	if err != nil {
		return nil, err
	}
	if status == http.StatusOK {
		return stream, nil
	}

	_ = stream.Close()
	message := "expected a complete attachment response"
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		message = "attachment access denied"
	}
	return nil, domain.NewUpstreamError(domain.ServiceWiki, "DownloadAttachment", status, "", message, "")
}
