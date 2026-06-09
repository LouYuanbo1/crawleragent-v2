package crawler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== NetworkResponse (UrlContent接口实现) ====================

func TestNetworkResponse_ImplementsUrlContent(t *testing.T) {
	var c UrlContent = &NetworkResponse{
		Url:        "https://example.com/api",
		UrlPattern: "/api/*",
		Body:       `{"key":"value"}`,
	}
	assert.Implements(t, (*UrlContent)(nil), c)
}

func TestNetworkResponse_Getters(t *testing.T) {
	resp := &NetworkResponse{
		Url:        "https://example.com/data",
		UrlPattern: "/data/*",
		Body:       `{"status":"ok"}`,
	}

	assert.Equal(t, "https://example.com/data", resp.GetUrl())
	assert.Equal(t, "/data/*", resp.GetUrlPattern())
	assert.Equal(t, []byte(`{"status":"ok"}`), resp.GetContent())
}

func TestNetworkResponse_Empty(t *testing.T) {
	resp := &NetworkResponse{}
	assert.Empty(t, resp.GetUrl())
	assert.Empty(t, resp.GetUrlPattern())
	assert.Empty(t, resp.GetContent())
}

// ==================== HtmlContent (UrlContent接口实现) ====================

func TestHtmlContent_ImplementsUrlContent(t *testing.T) {
	var c UrlContent = &HtmlContent{
		Url:     "https://example.com",
		Content: []byte("<html></html>"),
	}
	assert.Implements(t, (*UrlContent)(nil), c)
}

func TestHtmlContent_Getters(t *testing.T) {
	html := &HtmlContent{
		Url:     "https://example.com/page",
		Content: []byte("<body>hello</body>"),
	}

	assert.Equal(t, "https://example.com/page", html.GetUrl())
	assert.Empty(t, html.GetUrlPattern(), "HtmlContent.GetUrlPattern应返空")
	assert.Equal(t, []byte("<body>hello</body>"), html.GetContent())
}

func TestHtmlContent_Empty(t *testing.T) {
	html := &HtmlContent{}
	assert.Empty(t, html.GetUrl())
	assert.Empty(t, html.GetContent())
}

func TestHtmlContent_NilContent(t *testing.T) {
	html := &HtmlContent{Url: "about:blank"}
	assert.Empty(t, html.GetContent())
}

// ==================== NetworkConfig ====================

func TestNetworkConfig_ProcessFunc(t *testing.T) {
	captured := ""
	nc := &NetworkConfig{
		URLPattern: "/data/*",
		ProcessFunc: func(ctx context.Context, content UrlContent) error {
			captured = string(content.GetContent())
			return nil
		},
	}

	// 调用 ProcessFunc
	err := nc.ProcessFunc(context.Background(), &NetworkResponse{
		Url:        "https://example.com/data",
		UrlPattern: "/data/*",
		Body:       "processed",
	})
	assert.NoError(t, err)
	assert.Equal(t, "processed", captured)
}

func TestNetworkConfig_ProcessFuncError(t *testing.T) {
	sentinel := errors.New("custom error")
	nc := &NetworkConfig{
		URLPattern: "/api/*",
		ProcessFunc: func(ctx context.Context, content UrlContent) error {
			return sentinel
		},
	}

	err := nc.ProcessFunc(context.Background(), &NetworkResponse{})
	assert.ErrorIs(t, err, sentinel)
}

func TestNetworkConfig_EmptyURLPattern(t *testing.T) {
	nc := &NetworkConfig{URLPattern: ""}
	assert.Empty(t, nc.URLPattern)
	assert.Nil(t, nc.ProcessFunc)
}

// ==================== CrawlerParam ====================

func TestCrawlerParam_Struct(t *testing.T) {
	actions := []Action{
		&ClickAction{Selector: "#btn"},
	}
	configs := []*NetworkConfig{
		{URLPattern: "/api/*"},
	}

	param := &CrawlerParam{
		URL:            "https://example.com",
		NetworkConfigs: configs,
		Actions:        actions,
	}

	assert.Equal(t, "https://example.com", param.URL)
	assert.Len(t, param.NetworkConfigs, 1)
	assert.Len(t, param.Actions, 1)
}

func TestCrawlerParam_EmptyActions(t *testing.T) {
	param := &CrawlerParam{
		URL:            "https://example.com",
		NetworkConfigs: nil,
		Actions:        nil,
	}

	assert.Empty(t, param.Actions)
	assert.Empty(t, param.NetworkConfigs)
}

// ==================== UrlContent 多态 ====================

func TestUrlContent_PolymorphicSlice(t *testing.T) {
	contents := []UrlContent{
		&NetworkResponse{Url: "https://a.com", UrlPattern: "/a/*", Body: "bodyA"},
		&HtmlContent{Url: "https://b.com", Content: []byte("bodyB")},
	}

	assert.Len(t, contents, 2)
	assert.Equal(t, "https://a.com", contents[0].GetUrl())
	assert.Equal(t, "/a/*", contents[0].GetUrlPattern())
	assert.Equal(t, []byte("bodyA"), contents[0].GetContent())

	assert.Equal(t, "https://b.com", contents[1].GetUrl())
	assert.Empty(t, contents[1].GetUrlPattern())
	assert.Equal(t, []byte("bodyB"), contents[1].GetContent())
}
