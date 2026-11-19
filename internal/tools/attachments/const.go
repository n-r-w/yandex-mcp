package attachments

const (
	attachmentDirPerm   = 0o750    // Permissions for directories created by the writer.
	attachmentFilePerm  = 0o600    // Permissions for files created by the writer.
	emptyAllowlistLabel = "(none)" // Display label when no values apply.
	parentDirMarker     = ".."     // Path segment that refers to the parent directory.
)
