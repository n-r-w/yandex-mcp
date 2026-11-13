package oauthfile

import "errors"

var (
	errEmptyToken     = errors.New("OAuth token file is empty")
	errMalformedToken = errors.New("OAuth token file must hold one ASCII token without spaces")
)
