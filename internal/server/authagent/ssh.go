package authagent

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/n-r-w/yandex-mcp/internal/domain"
)

// runSSH executes one reverse-tunnel connection attempt using the user's OpenSSH configuration.
func runSSH(ctx context.Context, target, address string) error {
	//nolint:gosec // native ssh, fixed options, validated loopback address, destination after --; no shell
	cmd := exec.CommandContext(ctx, "ssh", "-N", "-T",
		"-o", "BatchMode=yes",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "StrictHostKeyChecking=yes",
		"-o", "ServerAliveInterval=15",
		"-o", "ServerAliveCountMax=3",
		"-o", "ConnectTimeout=10",
		"-R", address+":"+address, "--", target)
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	err := cmd.Run()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil && diagnostics.Len() > 0 {
		return fmt.Errorf("%w\n%s", err, diagnostics.String())
	}
	return err
}

// runTunnel restarts SSH after a disconnect and waits for process exit on cancellation.
func (s *Service) runTunnel(ctx context.Context, target, address string) {
	for ctx.Err() == nil {
		err := runSSH(ctx, target, address)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			_ = domain.LogError(ctx, "ssh tunnel", err)
		} else {
			slog.InfoContext(ctx, "SSH tunnel closed")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(sshRetryDelay):
		}
	}
}
