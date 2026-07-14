package protocol

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAuthenticator 用于测试的 Authenticator 实现。
type mockAuthenticator struct {
	authenticateFn func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error)
	returnFn       func(accountID string, conn net.Conn, lastSeqID uint32)
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
	return m.authenticateFn(ctx, accountID, skipDrain)
}

func (m *mockAuthenticator) Return(accountID string, conn net.Conn, lastSeqID uint32) {
	if m.returnFn != nil {
		m.returnFn(accountID, conn, lastSeqID)
	} else if conn != nil {
		_ = conn.Close()
	}
}

func TestPooledAuthenticator_UseInnerWhenNoPool(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	inner := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			return fc, false, 0, nil
		},
	}

	auth := NewPooledAuthenticator(inner, nil, "srv", "http")
	conn, borrowed, seqID, err := auth.Authenticate(context.Background(), "test1", false)
	require.NoError(t, err)
	assert.Same(t, fc, conn)
	assert.False(t, borrowed)
	assert.Equal(t, uint32(0), seqID)
}

func TestPooledAuthenticator_UsePoolWhenAvailable(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	pool := NewAccountConnectionPool()
	pool.AcceptConn("test1", "srv", fc)

	inner := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			t.Fatal("连接池可用时不应调用 inner authenticator")
			return nil, false, 0, nil
		},
	}

	auth := NewPooledAuthenticator(inner, pool, "srv", "http")
	conn, borrowed, seqID, err := auth.Authenticate(context.Background(), "test1", false)
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.True(t, borrowed)
	assert.Equal(t, uint32(0), seqID)
}

func TestPooledAuthenticator_PassAuthErrorThrough(t *testing.T) {
	inner := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			return nil, false, 0, &AuthError{Code: -600, Message: "登录失败"}
		},
	}

	auth := NewPooledAuthenticator(inner, nil, "srv", "http")
	_, _, _, err := auth.Authenticate(context.Background(), "test1", false)
	require.Error(t, err)
	assert.True(t, IsRateLimitError(err), "限流错误应能被识别")
}

func TestSendMessagesOnce_UsesAuthenticator(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	auth := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			return fc, false, 0, nil
		},
	}

	messages := []RecordMessage{}
	sent, err := sendMessagesOnce(accountRunOptions{
		ReplayOptions: ReplayOptions{
			Context:     context.Background(),
			ServerAddr:  "srv",
			Messages:    messages,
			RepeatCount: 1,
		},
		accountID: "test1",
		auth:      auth,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, sent)
}

func TestReplayer_SendMessagesWithRetry_UsesAuthenticator(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	auth := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			return fc, false, 0, nil
		},
	}

	cfg := RetryConfig{
		MaxRetries:     0,
		RetryInterval:  0,
		MinConcurrency: 1,
	}

	opts := ReplayOptions{
		ServerAddr:    "srv",
		HTTPAddr:      "http",
		OpenID:        "test",
		RangeStart:    1,
		RangeEnd:      1,
		Messages:      []RecordMessage{},
		RepeatCount:   1,
		Context:       context.Background(),
		RetryConfig:   cfg,
		Concurrency:   1,
		Authenticator: auth,
	}

	err := NewReplayer(opts).SendMessagesWithRetry()
	require.NoError(t, err)
}

func TestReplayer_SendMessagesWithRetry_Defaults(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	auth := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, accountID string, skipDrain bool) (net.Conn, bool, uint32, error) {
			return fc, false, 0, nil
		},
	}

	// 零值配置应被 normalize 填充为默认值
	opts := ReplayOptions{
		ServerAddr:    "srv",
		HTTPAddr:      "http",
		OpenID:        "test",
		RangeStart:    1,
		RangeEnd:      1,
		Messages:      []RecordMessage{},
		Authenticator: auth,
	}

	err := NewReplayer(opts).SendMessages()
	require.NoError(t, err)
}

func TestReplayer_SendMessagesWithRetry_UsesConnPool(t *testing.T) {
	fc := NewFakeConn()
	defer fc.Close()

	pool := NewAccountConnectionPool()
	pool.AcceptConn("test1", "srv", fc)

	// 没有注入 Authenticator 时，Replayer 应默认创建 HTTPAuthenticator，
	// 并用 PooledAuthenticator 包装它。这里我们直接验证 PooledAuthenticator 包装逻辑。
	opts := ReplayOptions{
		ServerAddr: "srv",
		HTTPAddr:   "http",
		OpenID:     "test",
		RangeStart: 1,
		RangeEnd:   1,
		Messages:   []RecordMessage{},
		Context:    context.Background(),
		ConnPool:   pool,
		RetryConfig: RetryConfig{
			MaxRetries:     0,
			RetryInterval:  0,
			MinConcurrency: 1,
		},
		Concurrency: 1,
	}

	err := NewReplayer(opts).SendMessagesWithRetry()
	require.NoError(t, err)
}
