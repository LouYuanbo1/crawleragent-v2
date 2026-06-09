package crawler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ==================== extractPatterns ====================

func TestExtractPatterns(t *testing.T) {
	tests := []struct {
		name     string
		configs  []*NetworkConfig
		expected []string
	}{
		{
			name:     "nil切片",
			configs:  nil,
			expected: []string{},
		},
		{
			name:     "空切片",
			configs:  []*NetworkConfig{},
			expected: []string{},
		},
		{
			name:     "单个有效配置",
			configs:  []*NetworkConfig{{URLPattern: "/api/*"}},
			expected: []string{"/api/*"},
		},
		{
			name: "多个配置",
			configs: []*NetworkConfig{
				{URLPattern: "/api/*"},
				{URLPattern: "/static/*"},
			},
			expected: []string{"/api/*", "/static/*"},
		},
		{
			name:     "包含nil元素",
			configs:  []*NetworkConfig{nil, {URLPattern: "/api/*"}, nil},
			expected: []string{"/api/*"},
		},
		{
			name:     "包含空URLPattern",
			configs:  []*NetworkConfig{{URLPattern: ""}, {URLPattern: "/api/*"}},
			expected: []string{"/api/*"},
		},
		{
			name:     "全部为nil或空",
			configs:  []*NetworkConfig{nil, {URLPattern: ""}},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPatterns(tt.configs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== collectErrors ====================

func TestCollectErrors(t *testing.T) {
	t.Run("空通道", func(t *testing.T) {
		ch := make(chan error)
		close(ch)
		err := collectErrors(ch)
		assert.NoError(t, err)
	})

	t.Run("单个错误", func(t *testing.T) {
		ch := make(chan error, 1)
		ch <- errors.New("test error")
		close(ch)
		err := collectErrors(ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
	})

	t.Run("多个错误", func(t *testing.T) {
		ch := make(chan error, 3)
		ch <- errors.New("err1")
		ch <- errors.New("err2")
		ch <- errors.New("err3")
		close(ch)
		err := collectErrors(ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "err1")
		assert.Contains(t, err.Error(), "err2")
		assert.Contains(t, err.Error(), "err3")
	})

	t.Run("errors.Join保留错误树", func(t *testing.T) {
		sentinel := io.EOF
		ch := make(chan error, 2)
		ch <- sentinel
		ch <- errors.New("other")
		close(ch)
		err := collectErrors(ch)
		assert.True(t, errors.Is(err, io.EOF), "errors.Join应保留子错误的可检测性")
	})
}

// ==================== context sleep ====================

func TestSleep(t *testing.T) {
	c := &poolCrawler{logger: slogDefault()}

	t.Run("正常等待完成", func(t *testing.T) {
		start := time.Now()
		err := c.sleep(context.Background(), 50*time.Millisecond)
		elapsed := time.Since(start)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
	})

	t.Run("context取消提前返回", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := c.sleep(ctx, time.Second)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestCheckCtx(t *testing.T) {
	c := &poolCrawler{logger: slogDefault()}

	t.Run("未取消", func(t *testing.T) {
		assert.NoError(t, c.checkCtx(context.Background()))
	})

	t.Run("已取消", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.ErrorIs(t, c.checkCtx(ctx), context.Canceled)
	})
}

// slogDefault 返回测试用的默认slog logger
func slogDefault() *slog.Logger {
	return slog.Default()
}
