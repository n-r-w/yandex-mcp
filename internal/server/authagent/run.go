package authagent

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

// Run serves loopback HTTP, maintains SSH, and joins both process lifecycles on shutdown.
func (s *Service) Run(ctx context.Context, port int, sshTarget string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	address, err := config.AuthAgentAddress(port)
	if err != nil {
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
	sshDone := make(chan struct{})
	go func() { defer close(sshDone); s.runTunnel(ctx, sshTarget, address) }()
	stopped := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = server.Close()
		case <-stopped:
		}
	}()
	err = server.Serve(listener)
	cancel()
	close(stopped)
	<-sshDone
	s.Close()
	if errors.Is(err, http.ErrServerClosed) && ctx.Err() != nil {
		return nil
	}
	return err
}
