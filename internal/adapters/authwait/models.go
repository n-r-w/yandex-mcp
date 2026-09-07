package authwait

import "context"

type acquisition struct {
	done    chan struct{}
	cancel  context.CancelFunc
	waiters int
	token   string
	err     error
}
