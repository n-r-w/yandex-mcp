package yc

//nolint:gosec // token syntax, not a credential
const tokenRegexPattern = `t1\.[A-Z0-9a-z_-]+={0,2}\.[A-Z0-9a-z_-]{86}={0,2}`
