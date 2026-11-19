package wiki

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// getAttachment validates delivery before downloading and returns only the requested result.
func (r *Registrator) getAttachment(ctx context.Context, input getAttachmentInput) (*getAttachmentOutput, error) {
	attachmentPath, fileName, err := parseAttachmentPath(input.AttachmentPath)
	if err != nil {
		return nil, err
	}

	savePath, err := r.prepareAttachmentDelivery(ctx, input, fileName)
	if err != nil {
		return nil, err
	}

	stream, err := r.adapter.DownloadAttachment(ctx, attachmentPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = stream.Close() }()

	if input.GetContent {
		return r.readAttachmentText(ctx, stream)
	}
	size, err := r.delivery.Write(savePath, input.Override, stream)
	if err != nil {
		return nil, domain.LogError(
			ctx,
			string(domain.ServiceWiki),
			errors.New("could not save complete Wiki attachment"),
		)
	}
	return &getAttachmentOutput{Content: nil, SavedPath: savePath, Size: size}, nil
}

// parseAttachmentPath validates the copied spelling without decoding or cleaning it.
func parseAttachmentPath(path string) (attachmentPath, fileName string, err error) {
	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")
	if len(segments) < 3 || segments[len(segments)-2] != ".files" || strings.Contains(segments[0], ":") ||
		strings.ContainsAny(path, "\\?#") || strings.ContainsFunc(path, unicode.IsControl) {
		return "", "", errAttachmentPath
	}
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" || segment == "." || segment == ".." {
			return "", "", errAttachmentPath
		}
	}
	return path, segments[len(segments)-1], nil
}

// prepareAttachmentDelivery applies the policy for the selected mode without writing to disk.
func (r *Registrator) prepareAttachmentDelivery(
	ctx context.Context,
	input getAttachmentInput,
	fileName string,
) (string, error) {
	savePath := strings.TrimSpace(input.SavePath)
	if input.GetContent == (savePath != "") {
		return "", errors.New("choose exactly one of get_content: true or save_path")
	}
	if savePath != input.SavePath {
		return "", errors.New("save_path must not have leading or trailing whitespace")
	}
	if input.GetContent {
		return "", r.delivery.ValidateViewExtension(fileName, "attachment_path filename")
	}
	return r.delivery.PrepareSavePath(ctx, domain.ServiceWiki, savePath, input.Override)
}

// readAttachmentText returns all bytes as UTF-8 text only when they fit the configured limit.
func (r *Registrator) readAttachmentText(ctx context.Context, stream io.Reader) (*getAttachmentOutput, error) {
	limited := &io.LimitedReader{R: stream, N: r.inlineMaxBytes}
	data, err := io.ReadAll(limited)
	if err == nil && limited.N == 0 {
		// Probe separately so the limit need not be incremented at math.MaxInt64.
		var extra [1]byte
		var n int
		n, err = io.ReadFull(stream, extra[:])
		if n > 0 {
			return nil, fmt.Errorf(
				"attachment exceeds inline limit of %d bytes; use save_path to save it",
				r.inlineMaxBytes,
			)
		}
		if errors.Is(err, io.EOF) {
			err = nil
		}
	}
	if err != nil {
		return nil, domain.LogError(
			ctx,
			string(domain.ServiceWiki),
			errors.New("could not read complete Wiki attachment"),
		)
	}
	if !utf8.Valid(data) {
		return nil, errors.New("attachment is not valid UTF-8; use save_path to save it")
	}
	content := string(data)
	return &getAttachmentOutput{Content: &content, SavedPath: "", Size: int64(len(data))}, nil
}
