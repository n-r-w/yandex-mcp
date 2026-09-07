package authagent

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Run serves loopback HTTP until cancellation and waits for token-source cleanup.
func (s *Service) Run(ctx context.Context, address string) error {
	if err := config.ValidateLoopbackAddress(address); err != nil {
		return err
	}
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	//nolint:exhaustruct_v5 // header timeout does not limit interactive authentication
	server := &http.Server{
		Handler:           s,
		ReadHeaderTimeout: headerTimeout,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = server.Close()
		case <-stopped:
		}
	}()
	err = server.Serve(listener)
	close(stopped)
	s.Close()
	if errors.Is(err, http.ErrServerClosed) && ctx.Err() != nil {
		return nil
	}
	return err
}
