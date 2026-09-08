package authagent

import "time"

const (
	headerTimeout   = 5 * time.Second
	sshRetryDelay   = 5 * time.Second
	maxRequestBytes = 4096
)
