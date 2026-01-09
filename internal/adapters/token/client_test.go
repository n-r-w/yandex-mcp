package token

import (
	"context"
	"errors"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/n-r-w/yandex-mcp/internal/config"
)

func testConfig(refreshPeriod time.Duration) *config.Config {
	//nolint:exhaustruct // only IAMTokenRefreshPeriod relevant for token tests
	return &config.Config{IAMTokenRefreshPeriod: refreshPeriod}
}

// makeValidToken creates a token string that matches the required regex pattern.
// The token format is: t1.<prefix>.<86-char-suffix>[padding].
func makeValidToken(name string) string {
	// Create a simple distinguishable prefix
	prefix := "prefix_" + name
	// Create exactly 86 characters for the suffix (padded with repeated name if needed)
	suffix := name
	for len(suffix) < 86 {
		suffix += name
	}
	suffix = suffix[:86]
	return "t1." + prefix + "." + suffix
}

// atomicTime provides a thread-safe mutable time value for tests.
type atomicTime struct {
	mu  sync.RWMutex
	now time.Time
}

func newAtomicTime(t time.Time) *atomicTime {
	return &atomicTime{mu: sync.RWMutex{}, now: t}
}

func (a *atomicTime) Now() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.now
}

func (a *atomicTime) Advance(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.now = a.now.Add(d)
}

func TestProvider_Token_CachesBetweenCalls(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(makeValidToken("test123")), nil).
		Times(1)

	ctx := context.Background()

	tok1, err1 := provider.Token(ctx)
	require.NoError(t, err1)
	assert.Equal(t, makeValidToken("test123"), tok1)

	tok2, err2 := provider.Token(ctx)
	require.NoError(t, err2)
	assert.Equal(t, makeValidToken("test123"), tok2)
}

func TestProvider_Token_RefreshesAfterPeriodExpires(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	clock := newAtomicTime(time.Now())
	provider.setNowFunc(clock.Now)

	gomock.InOrder(
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("v1")), nil),
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("v2")), nil),
	)

	ctx := context.Background()

	tok1, err1 := provider.Token(ctx)
	require.NoError(t, err1)
	assert.Equal(t, makeValidToken("v1"), tok1)

	clock.Advance(time.Hour + time.Minute)

	tok2, err2 := provider.Token(ctx)
	require.NoError(t, err2)
	assert.Equal(t, makeValidToken("v2"), tok2)
}

func TestProvider_ForceRefresh_BypassesCache(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	gomock.InOrder(
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("cached")), nil),
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("forced")), nil),
	)

	ctx := context.Background()

	tok1, err1 := provider.Token(ctx)
	require.NoError(t, err1)
	assert.Equal(t, makeValidToken("cached"), tok1)

	tok2, err2 := provider.ForceRefresh(ctx)
	require.NoError(t, err2)
	assert.Equal(t, makeValidToken("forced"), tok2)
}

func TestProvider_ForceRefresh_UpdatesCacheForSubsequentCalls(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	gomock.InOrder(
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("original")), nil),
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("refreshed")), nil),
	)

	ctx := context.Background()

	_, _ = provider.Token(ctx)
	_, _ = provider.ForceRefresh(ctx)

	tok, err := provider.Token(ctx)
	require.NoError(t, err)
	assert.Equal(t, makeValidToken("refreshed"), tok)
}

func TestProvider_Token_ConcurrentCallsTriggerSingleExecution(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	var execCount atomic.Int32
	execStarted := make(chan struct{})
	execProceed := make(chan struct{})

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		DoAndReturn(func(_ context.Context, _ string, _ ...string) ([]byte, error) {
			execCount.Add(1)
			close(execStarted)
			<-execProceed
			return []byte(makeValidToken("concurrent")), nil
		}).
		Times(1)

	ctx := context.Background()
	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	tokens := make([]string, goroutines)
	errs := make([]error, goroutines)

	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			tokens[idx], errs[idx] = provider.Token(ctx)
		}(i)
	}

	<-execStarted
	close(execProceed)
	wg.Wait()

	for i := range goroutines {
		require.NoError(t, errs[i], "goroutine %d should not error", i)
		assert.Equal(t, makeValidToken("concurrent"), tokens[i], "goroutine %d should get the same token", i)
	}

	assert.Equal(t, int32(1), execCount.Load(), "yc should be executed exactly once")
}

func TestProvider_ForceRefresh_ErrorDoesNotContainToken(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	tokenV1 := makeValidToken("secret")
	sensitiveOutput := "partial-token-" + tokenV1 + " remaining output"
	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(tokenV1), nil)
	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return(nil, errors.New(sensitiveOutput))

	ctx := context.Background()
	_, _ = provider.Token(ctx)
	_, err := provider.ForceRefresh(ctx)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), tokenV1, "error must not leak previously cached token")
}

func TestProvider_Token_ErrorDoesNotLeakOutputWhenTokenNotFound(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	// Output with potentially sensitive data but no valid token
	sensitiveOutput := `[ERROR] Authentication failed for user: admin@example.com
[DEBUG] API Key: ak_1234567890abcdef
[INFO] Session ID: sess_xyz789
Status: unauthorized`

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(sensitiveOutput), nil)

	ctx := context.Background()
	_, err := provider.Token(ctx)

	require.Error(t, err)
	require.ErrorIs(t, err, errTokenNotFound)
	assert.NotContains(t, err.Error(), "admin@example.com", "error must not leak stdout")
	assert.NotContains(t, err.Error(), "ak_1234567890abcdef", "error must not leak stdout")
	assert.NotContains(t, err.Error(), "sess_xyz789", "error must not leak stdout")
	assert.NotContains(t, err.Error(), "unauthorized", "error must not leak stdout")
}

func TestProvider_Token_RegexPatternMatches(t *testing.T) {
	t.Parallel()

	token := makeValidToken("test123")
	re := regexp.MustCompile(tokenRegexPattern)
	match := re.Find([]byte(token))

	require.NotNil(t, match, "regex should match valid token")
	assert.Equal(t, token, string(match))
}

func TestProvider_Token_ExtractsTokenWithRegex_CleanOutput(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	validToken := makeValidToken("clean")
	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(validToken), nil)

	ctx := context.Background()
	tok, err := provider.Token(ctx)

	require.NoError(t, err)
	assert.Equal(t, validToken, tok)
}

func TestProvider_Token_ExtractsTokenWithRegex_NoisyOutput(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	validToken := makeValidToken("noisy")
	noisyOutput := `[2024-01-07 10:15:30] INFO: Initializing Yandex Cloud CLI
[2024-01-07 10:15:31] DEBUG: Checking credentials
[2024-01-07 10:15:32] INFO: Token created: ` + validToken + `
[2024-01-07 10:15:32] INFO: Token expires in 12 hours
Additional log data and metadata here...`

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(noisyOutput), nil)

	ctx := context.Background()
	tok, err := provider.Token(ctx)

	require.NoError(t, err)
	assert.Equal(t, validToken, tok)
}

func TestProvider_Token_ExtractsTokenWithRegex_WhitespaceAndNewlines(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	// Token with exactly 86 chars in suffix + "=" padding
	validToken := makeValidToken("whitespace") + "="
	outputWithWhitespace := "\n\n  " + validToken + "  \n\t\n"

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(outputWithWhitespace), nil)

	ctx := context.Background()
	tok, err := provider.Token(ctx)

	require.NoError(t, err)
	assert.Equal(t, validToken, tok)
}

func TestProvider_Token_ReturnsErrorWhenNoTokenMatch(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte("some random output without a valid token format"), nil)

	ctx := context.Background()
	_, err := provider.Token(ctx)

	require.Error(t, err)
	require.ErrorIs(t, err, errTokenNotFound)
	assert.NotContains(t, err.Error(), "random output", "error must not leak yc output")
}

func TestProvider_Token_ReturnsErrorOnInvalidTokenFormat(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	// Token-like but invalid format (too short suffix)
	invalidToken := "t1.prefix.short"
	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(invalidToken), nil)

	ctx := context.Background()
	_, err := provider.Token(ctx)

	require.Error(t, err)
	require.ErrorIs(t, err, errTokenNotFound)
}

func TestProvider_Token_ReturnsErrorOnEmptyOutput(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte(""), nil)

	ctx := context.Background()
	_, err := provider.Token(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, errEmptyToken)
}

func TestProvider_Token_ReturnsErrorOnWhitespaceOnlyOutput(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return([]byte("   \n\t"), nil)

	ctx := context.Background()
	_, err := provider.Token(ctx)

	require.Error(t, err)
	require.ErrorIs(t, err, errTokenNotFound)
}

func TestProvider_Token_ContextCancellationPropagates(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		Return(nil, context.Canceled)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Token(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestProvider_Token_ConcurrentCallsDuringRefreshShareResult(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	clock := newAtomicTime(time.Now())
	provider.setNowFunc(clock.Now)

	var refreshCount atomic.Int32
	refreshStarted := make(chan struct{})
	refreshProceed := make(chan struct{})

	gomock.InOrder(
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			Return([]byte(makeValidToken("initial")), nil),
		mockExec.EXPECT().
			Execute(gomock.Any(), "yc", "iam", "create-token").
			DoAndReturn(func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				refreshCount.Add(1)
				close(refreshStarted)
				<-refreshProceed
				return []byte(makeValidToken("refreshed")), nil
			}),
	)

	ctx := context.Background()

	_, _ = provider.Token(ctx)

	clock.Advance(time.Hour + time.Minute)

	const goroutines = 5
	var wg sync.WaitGroup
	wg.Add(goroutines)

	tokens := make([]string, goroutines)
	for i := range goroutines {
		go func(idx int) {
			defer wg.Done()
			tokens[idx], _ = provider.Token(ctx)
		}(i)
	}

	<-refreshStarted
	close(refreshProceed)
	wg.Wait()

	for i := range goroutines {
		assert.Equal(t, makeValidToken("refreshed"), tokens[i])
	}
	assert.Equal(t, int32(1), refreshCount.Load(), "refresh should happen exactly once")
}

func TestProvider_Token_WaiterReceivesRefreshError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	mockExec := NewMockICommandExecutor(ctrl)
	cfg := testConfig(time.Hour)
	provider := NewProvider(cfg)
	provider.setExecutor(mockExec)

	refreshStarted := make(chan struct{})
	refreshProceed := make(chan struct{})

	mockExec.EXPECT().
		Execute(gomock.Any(), "yc", "iam", "create-token").
		DoAndReturn(func(_ context.Context, _ string, _ ...string) ([]byte, error) {
			close(refreshStarted)
			<-refreshProceed
			return nil, errTokenFetchFailed
		}).
		Times(1)

	ctx := context.Background()

	// Start leader
	leaderDone := make(chan error, 1)
	go func() {
		_, err := provider.Token(ctx)
		leaderDone <- err
	}()

	<-refreshStarted

	// At this point, leader is in Execute, inflight=true, channel exists.
	// Start waiter which will enter Token(), acquire lock, see inflight=true,
	// save channel reference, release lock, and block on channel.
	waiterStarted := make(chan struct{})
	waiterDone := make(chan error, 1)
	go func() {
		close(waiterStarted)
		_, err := provider.Token(ctx)
		waiterDone <- err
	}()

	// Wait for waiter goroutine to start
	<-waiterStarted

	// The waiter needs to acquire the mutex and start waiting on the channel.
	// Since the mutex is not held by leader (leader released it before calling Execute),
	// the waiter will quickly acquire it, see inflight=true, and block on channel.
	// Yield to give the waiter time to reach the channel wait.
	runtime.Gosched()

	close(refreshProceed)

	leaderErr := <-leaderDone
	waiterErr := <-waiterDone

	require.Error(t, leaderErr)
	require.ErrorIs(t, leaderErr, errTokenFetchFailed)

	require.Error(t, waiterErr)
	require.ErrorIs(t, waiterErr, errTokenFetchFailed)
}
