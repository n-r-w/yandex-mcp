package yc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMain makes the native test executable act as yc in child processes.
func TestMain(m *testing.M) {
	if os.Getenv("YANDEX_MCP_TEST_YC") == "1" {
		if len(os.Args) == 3 && os.Args[1] == "iam" && os.Args[2] == "create-token" {
			_, _ = fmt.Fprintln(os.Stdout, "t1.prefix."+strings.Repeat("a", 86))
			os.Exit(0)
		}
		if len(os.Args) != 5 || os.Args[1] != "iam" || os.Args[2] != "create-token" || os.Args[3] != "--profile" {
			os.Exit(2)
		}
		switch os.Args[4] {
		case "wait":
			if err := os.WriteFile(os.Getenv("YANDEX_MCP_TEST_READY"), []byte("ready"), 0o600); err != nil {
				os.Exit(3)
			}
			for {
				time.Sleep(time.Hour)
			}
		case "fail":
			fmt.Fprintln(os.Stderr, "private-login-url secret-token")
			os.Exit(1)
		case "invalid":
			_, _ = fmt.Fprint(os.Stdout, "private-login-url")
			os.Exit(0)
		case "empty":
			os.Exit(0)
		default:
			_, _ = fmt.Fprint(os.Stdout, "login information\n"+"t1.prefix."+strings.Repeat("a", 86)+"\n")
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// Native subprocess contract: fixed arguments, token parsing, safe errors, and
// cancellation. The child test executable replaces yc; no shell or network is used.
func TestAcquire(t *testing.T) {
	t.Setenv("YANDEX_MCP_TEST_YC", "1")
	executable, err := os.Executable()
	require.NoError(t, err)
	source := New(executable)
	token, err := source.Acquire(t.Context(), "work")
	require.NoError(t, err)
	require.Equal(t, "t1.prefix."+strings.Repeat("a", 86), token)
	token, err = source.Acquire(t.Context(), "")
	require.NoError(t, err)
	require.Equal(t, "t1.prefix."+strings.Repeat("a", 86), token)
	for _, profile := range []string{"fail", "invalid", "empty"} {
		token, err = source.Acquire(t.Context(), profile)
		require.Error(t, err)
		require.Empty(t, token)
		switch profile {
		case "fail":
			require.Contains(t, err.Error(), "exit status 1")
			require.Contains(t, err.Error(), "private-login-url secret-token")
		case "invalid":
			require.ErrorContains(t, err, "token not found in yc output")
		case "empty":
			require.ErrorContains(t, err, "empty token received from yc")
		}
	}
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("YANDEX_MCP_TEST_READY", ready)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, acquireErr := source.Acquire(ctx, "wait"); result <- acquireErr }()
	require.Eventually(
		t,
		func() bool { _, statErr := os.Stat(ready); return statErr == nil },
		5*time.Second,
		time.Millisecond,
	)
	cancel()
	select {
	case err = <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("yc did not exit after cancellation")
	}
	token, err = source.Acquire(t.Context(), "work")
	require.NoError(t, err)
	require.NotEmpty(t, token)
}
