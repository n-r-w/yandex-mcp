package oauthfile

import "errors"

// Errors of token file reading; Acquire wraps each in domain.AuthenticationError.
var (
	// errEmptyToken reports a token file without a token.
	errEmptyToken = errors.New("OAuth token file is empty")
	// errMalformedToken reports a token file whose content is not one token.
	errMalformedToken = errors.New("OAuth token file must hold one ASCII token without spaces")
	// errUnknownProfile reports a profile without a file in the profiles directory.
	errUnknownProfile = errors.New("unknown OAuth profile")
	// errInvalidProfile reports a profile name that could address a file outside the profiles directory.
	errInvalidProfile = errors.New("invalid OAuth profile name")
	// errNoProfilesDir reports an auth-agent without a resolved profiles directory.
	errNoProfilesDir = errors.New("OAuth profiles directory is not configured")
)
