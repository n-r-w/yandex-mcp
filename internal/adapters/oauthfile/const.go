package oauthfile

const (
	// byteOrderMark is the UTF-8 encoding of U+FEFF.
	byteOrderMark = "\xef\xbb\xbf"

	// profilePattern keeps a profile name inside the profiles directory: no separators, no leading dot.
	profilePattern = `^[A-Za-z0-9][A-Za-z0-9._-]*$`
)
