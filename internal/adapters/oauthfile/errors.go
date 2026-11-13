package oauthfile

import "errors"

var (
	errEmptyToken     = errors.New("OAuth token file is empty")
	errMalformedToken = errors.New("OAuth token file must hold one ASCII token without spaces")
	errUnknownProfile = errors.New("unknown OAuth profile")
	errInvalidProfile = errors.New("invalid OAuth profile name")
	errNoProfilesDir  = errors.New("OAuth profiles directory is not configured")
)
