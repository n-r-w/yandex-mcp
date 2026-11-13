package authagent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
	target := os.Args[len(os.Args)-1]
	allowedTargets := strings.Split(os.Getenv("YANDEX_MCP_TEST_SSH_TARGETS"), ",")
	if len(allowedTargets) == 1 && allowedTargets[0] == "" {
		allowedTargets = []string{"test-server"}
	}
	if !slices.Contains(allowedTargets, target) {
		_, _ = fmt.Fprintln(os.Stderr, "unexpected SSH target")
		os.Exit(2)
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
		target,
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
		if hold := os.Getenv("YANDEX_MCP_TEST_SSH_HOLD"); hold != "" {
			recordTargetAndHold(file, target, hold)
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

// recordTargetAndHold records the started target and keeps its SSH running only when hold lists it.
func recordTargetAndHold(file *os.File, target, hold string) {
	_, _ = fmt.Fprintln(file, target)
	_ = file.Close()
	if !slices.Contains(strings.Split(hold, ","), target) {
		os.Exit(7)
	}
	for {
		time.Sleep(time.Hour)
	}
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

// TestMultipleTunnels checks that all SSH targets run concurrently and stop after cancellation.
func TestMultipleTunnels(t *testing.T) {
	installSSHFixture(t)
	events := filepath.Join(t.TempDir(), "events")
	t.Setenv("YANDEX_MCP_TEST_SSH_EVENTS", events)
	t.Setenv("YANDEX_MCP_TEST_SSH_HOLD", "first-server,second-server")
	t.Setenv("YANDEX_MCP_TEST_SSH_TARGETS", "first-server,second-server")
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	var service Service
	go func() {
		defer close(done)
		service.runTunnels(ctx, []string{"first-server", "second-server"}, "127.0.0.1:18765")
	}()
	defer func() {
		cancel()
		<-done
	}()
	require.Eventually(t, func() bool {
		data, _ := os.ReadFile(events)
		targets := strings.Fields(string(data))
		return slices.Contains(targets, "first-server") && slices.Contains(targets, "second-server")
	}, 5*time.Second, 25*time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "SSH processes did not stop")
	}
}

// TestTunnelRestartIsIndependent checks that a failed tunnel restarts while the tunnel to another target keeps running.
func TestTunnelRestartIsIndependent(t *testing.T) {
	installSSHFixture(t)
	events := filepath.Join(t.TempDir(), "events")
	t.Setenv("YANDEX_MCP_TEST_SSH_EVENTS", events)
	t.Setenv("YANDEX_MCP_TEST_SSH_HOLD", "second-server")
	t.Setenv("YANDEX_MCP_TEST_SSH_TARGETS", "first-server,second-server")
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	var service Service
	go func() {
		defer close(done)
		service.runTunnels(ctx, []string{"first-server", "second-server"}, "127.0.0.1:18765")
	}()
	defer func() {
		cancel()
		<-done
	}()
	countStarts := func(target string) int {
		data, _ := os.ReadFile(events)
		return strings.Count(string(data), target+"\n")
	}
	require.Eventually(t, func() bool { return countStarts("first-server") >= 2 }, 3*sshRetryDelay, 25*time.Millisecond)
	require.Equal(t, 1, countStarts("second-server"))
}

// TestRunSSH checks native PATH lookup, fixed forwarding arguments, and original stderr.
func TestRunSSH(t *testing.T) {
	t.Setenv("YANDEX_MCP_TEST_SSH_EVENTS", "")
	installSSHFixture(t)
	err := runSSH(t.Context(), "test-server", "127.0.0.1:18765")
	require.ErrorContains(t, err, "exit status 7")
	require.ErrorContains(t, err, "original SSH diagnostic")
}
