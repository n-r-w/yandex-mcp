package wiki

import "errors"

// errAttachmentPath describes the required attachment link shape.
var errAttachmentPath = errors.New(
	"attachment_path must be a Wiki attachment path: slug/.files/filename with an optional initial slash",
)
