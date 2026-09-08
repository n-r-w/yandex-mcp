package authagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMain runs the native test executable as an SSH fixture in child processes.
func TestMain(m *testing.M) {
	if os.Getenv("YANDEX_MCP_TEST_SSH") != "1" {
		os.Exit(m.Run())
	}
	expected := []string{
		"-N",
		"-T",
		"-o",
		"BatchMode=yes",
		"-o",
		"ExitOnForwardFailure=yes",
		"-o",
		"StrictHostKeyChecking=yes",
		"-o",
		"ServerAliveInterval=15",
		"-o",
		"ServerAliveCountMax=3",
		"-o",
		"ConnectTimeout=10",
		"-R",
		"127.0.0.1:18765:127.0.0.1:18765",
		"--",
		"test-server",
	}
	if strings.Join(os.Args[1:], "\n") != strings.Join(expected, "\n") {
		_, _ = fmt.Fprintln(os.Stderr, "unexpected SSH arguments")
		os.Exit(2)
	}
	if path := os.Getenv("YANDEX_MCP_TEST_SSH_EVENTS"); path != "" {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			os.Exit(3)
		}
		_, _ = fmt.Fprintln(file, os.Getpid())
		_ = file.Close()
		contents, _ := os.ReadFile(path)
		if len(strings.Fields(string(contents))) > 1 {
			for {
				time.Sleep(time.Hour)
			}
		}
	}
	_, _ = fmt.Fprintln(os.Stderr, "original SSH diagnostic")
	os.Exit(7)
}

// installSSHFixture puts a native executable on PATH without touching user settings.
func installSSHFixture(t *testing.T) {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	data, err := os.ReadFile(executable)
	require.NoError(t, err)
	name := "ssh"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	directory := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(directory, name), data, 0o700))
	t.Setenv("PATH", directory)
	t.Setenv("YANDEX_MCP_TEST_SSH", "1")
}

// TestTunnelRestartAndCancellation checks restart after failure and joined shutdown of a live SSH process.
func TestTunnelRestartAndCancellation(t *testing.T) {
	installSSHFixture(t)
	events := filepath.Join(t.TempDir(), "events")
	t.Setenv("YANDEX_MCP_TEST_SSH_EVENTS", events)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	var service Service
	done := make(chan struct{})
	go func() { defer close(done); service.runTunnel(ctx, "test-server", "127.0.0.1:18765") }()
	defer func() { cancel(); <-done }()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			require.FailNow(t, "tunnel stopped before cancellation")
		case <-ctx.Done():
			require.FailNow(t, "SSH did not restart", ctx.Err().Error())
		case <-ticker.C:
			data, _ := os.ReadFile(events)
			if len(strings.Fields(string(data))) < 2 {
				continue
			}
			cancel()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				require.FailNow(t, "SSH process did not stop")
			}
			return
		}
	}
}

// TestRunSSH checks native PATH lookup, fixed forwarding arguments, and original stderr.
func TestRunSSH(t *testing.T) {
	t.Setenv("YANDEX_MCP_TEST_SSH_EVENTS", "")
	installSSHFixture(t)
	err := runSSH(t.Context(), "test-server", "127.0.0.1:18765")
	require.ErrorContains(t, err, "exit status 7")
	require.ErrorContains(t, err, "original SSH diagnostic")
}
