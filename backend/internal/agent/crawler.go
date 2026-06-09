package agent

import (
	"context"
	"crawleragent-v2/config"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/go-shiori/go-readability"
)

type Crawler interface {
	Close() error
	ClosePage() error
	Navigate(url string) error
	DoClick(selector string) error
	DoClickX(selector string) error
	DoScroll(scrollY int) error
	DoJavaScript(ctx context.Context, js string, args ...any) (string, error)
	GetHTML() (string, error)
	GetArticle() (string, error)
}

type crawler struct {
	browser *rod.Browser
	page    *rod.Page
	logger  *slog.Logger
}

func InitBrowserCrawler(cfg *config.RodConfig) (Crawler, error) {
	logger := slog.Default().With("component", "crawler")

	dataDir := fmt.Sprintf("%s/instance_0", cfg.UserDataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建实例数据目录失败: %v", err)
	}

	launcher := CreateLauncher(cfg.UserMode,
		WithBin(cfg.Bin),
		WithUserDataDir(dataDir),
		WithHeadless(cfg.Headless),
		WithDisableBlinkFeatures(cfg.DisableBlinkFeatures),
		WithIncognito(cfg.Incognito),
		WithDisableDevShmUsage(cfg.DisableDevShmUsage),
		WithNoSandbox(cfg.NoSandbox),
		WithUserAgent(cfg.UserAgent),
		WithLeakless(cfg.Leakless),
		WithDisableBackgroundNetworking(cfg.DisableBackgroundNetworking),
		WithDisableBackgroundTimerThrottling(cfg.DisableBackgroundTimerThrottling),
		WithRemoteDebuggingPort(cfg.BasicRemoteDebuggingPort),
	)
	url, err := launcher.Launch()
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %v", err)
	}
	logger.Info("浏览器就绪", "instance", 1, "url", url)

	browser := rod.New().ControlURL(url).Trace(cfg.Trace)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %v", err)
	}

	page := stealth.MustPage(browser)
	/*
		err = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:  cfg.DefaultPageWidth,
			Height: cfg.DefaultPageHeight,
		})
	*/

	return &crawler{
		browser: browser,
		page:    page,
		logger:  logger,
	}, nil
}

// Close 关闭爬虫，释放所有浏览器资源
func (c *crawler) Close() error {
	c.logger.Info("关闭爬虫")
	if err := c.browser.Close(); err != nil {
		return fmt.Errorf("关闭浏览器失败: %v", err)
	}
	return nil
}

func (c *crawler) ClosePage() error {
	if err := c.page.Close(); err != nil {
		return fmt.Errorf("关闭页面失败: %v", err)
	}
	return nil
}

// Navigate 导航到目标URL并等待页面加载与网络空闲
func (c *crawler) Navigate(url string) error {
	if err := c.page.Navigate(url); err != nil {
		return fmt.Errorf("导航失败: %v", err)
	}
	if err := c.page.WaitLoad(); err != nil {
		if strings.Contains(err.Error(), "navigated or closed") {
			return nil
		}
		return fmt.Errorf("页面加载等待失败: %v", err)
	}
	c.page.WaitRequestIdle(500*time.Millisecond, nil, nil, nil)
	return nil
}

// DoClick CSS选择器点击，点击后等待网络空闲
func (c *crawler) DoClick(selector string) error {
	el, err := c.page.Element(selector)
	if err != nil {
		return fmt.Errorf("未找到元素 %q: %v", selector, err)
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	c.page.WaitRequestIdle(500*time.Millisecond, nil, nil, nil)
	return nil
}

// DoClickX XPath点击，点击后等待网络空闲
func (c *crawler) DoClickX(selector string) error {
	el, err := c.page.ElementX(selector)
	if err != nil {
		return fmt.Errorf("未找到XPath元素 %q: %v", selector, err)
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return err
	}
	c.page.WaitRequestIdle(500*time.Millisecond, nil, nil, nil)
	return nil
}

// DoScroll 页面滚动（瞬时完成），滚动后等待网络空闲以捕获懒加载内容
func (c *crawler) DoScroll(scrollY int) error {
	_, err := c.page.Eval(`(dy) => window.scrollBy({top: dy, behavior: 'smooth'})`, scrollY)
	if err != nil {
		return err
	}
	c.page.WaitRequestIdle(500*time.Millisecond, nil, nil, nil)
	return nil
}

// doJavaScript 执行JS并回调处理结果
func (c *crawler) DoJavaScript(ctx context.Context, js string, args ...any) (string, error) {
	var result *proto.RuntimeRemoteObject
	var err error
	if len(args) > 0 {
		result, err = c.page.Eval(js, args...)
	} else {
		result, err = c.page.Eval("() => { " + js + " }")
	}
	if err != nil {
		return "", fmt.Errorf("JS执行失败: %v", err)
	}

	data, err := result.Value.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}

	return string(data), nil
}

func (c *crawler) GetHTML() (string, error) {
	return c.page.HTML()
}

func (c *crawler) GetArticle() (string, error) {
	html, err := c.page.HTML()
	if err != nil {
		return "", err
	}
	info, err := c.page.Info()
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(info.URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析 URL 失败: %v\n", err)
		return "", err
	}
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		return "", err
	}
	return article.TextContent, nil
}
