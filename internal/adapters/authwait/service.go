// Package authwait coordinates cancellable token acquisitions by profile.
package authwait

import (
	"context"
	"sync"
)

// Group owns active acquisitions. Its zero value is ready for use.
type Group struct {
	mu      sync.Mutex
	calls   map[string]*acquisition
	closed  bool
	workers sync.WaitGroup
}

// Do waits independently for an acquisition shared with callers of the same key.
func (g *Group) Do(ctx context.Context, key string, fn func(context.Context) (string, error)) (string, error) {
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		g.mu.Lock()
		if g.closed {
			g.mu.Unlock()
			return "", context.Canceled
		}
		if g.calls == nil {
			g.calls = make(map[string]*acquisition)
		}
		call := g.calls[key]
		if call != nil && call.waiters == 0 {
			g.mu.Unlock()
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-call.done:
				continue
			}
		}
		if call == nil {
			shared, cancel := context.WithCancel(context.WithoutCancel(ctx))
			call = &acquisition{done: make(chan struct{}), cancel: cancel, waiters: 0, token: "", err: nil}
			g.calls[key] = call
			g.workers.Add(1)
			go g.acquire(shared, key, call, fn)
		}
		call.waiters++
		g.mu.Unlock()
		return g.wait(ctx, call)
	}
}

func (g *Group) acquire(ctx context.Context, key string, call *acquisition, fn func(context.Context) (string, error)) {
	defer g.workers.Done()
	defer call.cancel()
	token, err := fn(ctx)
	g.mu.Lock()
	call.token, call.err = token, err
	delete(g.calls, key)
	close(call.done)
	g.mu.Unlock()
}

func (g *Group) wait(ctx context.Context, call *acquisition) (string, error) {
	select {
	case <-ctx.Done():
		g.mu.Lock()
		call.waiters--
		if call.waiters == 0 {
			call.cancel()
		}
		g.mu.Unlock()
		return "", ctx.Err()
	case <-call.done:
		if err := ctx.Err(); err != nil {
			return "", err
		}
		return call.token, call.err
	}
}

// Close cancels all acquisitions and waits for their cleanup. It is idempotent.
func (g *Group) Close() {
	g.mu.Lock()
	g.closed = true
	for _, call := range g.calls {
		call.cancel()
	}
	g.mu.Unlock()
	g.workers.Wait()
}
