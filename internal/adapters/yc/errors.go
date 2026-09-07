package yc

import "errors"

var (
	errEmptyToken    = errors.New("empty token received from yc")
	errTokenNotFound = errors.New("token not found in yc output")
)
