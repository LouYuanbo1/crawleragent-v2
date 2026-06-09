package crawler

import (
	"context"
	"crawleragent-v2/config"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Crawler 爬虫接口
type Crawler interface {
	Crawl(ctx context.Context, params []*CrawlerParam) error
	Close()
}

// poolCrawler 基于浏览器池的爬虫实现
type poolCrawler struct {
	pool        rod.Pool[rod.Browser]
	poolSize    int
	config      *config.RodConfig
	browserURLs []string
	mu          sync.Mutex
	logger      *slog.Logger
}

// InitBrowserPoolCrawler 创建浏览器池爬虫
func InitBrowserPoolCrawler(cfg *config.RodConfig, size int) (Crawler, error) {
	logger := slog.Default().With("component", "crawler")
	urls := make([]string, 0, size)

	for id := range size {
		dataDir := fmt.Sprintf("%s/instance_%d", cfg.UserDataDir, id)
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
			WithRemoteDebuggingPort(cfg.BasicRemoteDebuggingPort+id),
		)
		urlStr, err := launcher.Launch()
		if err != nil {
			return nil, fmt.Errorf("启动浏览器失败: %v", err)
		}
		logger.Info("浏览器就绪", "instance", id, "url", urlStr)
		urls = append(urls, urlStr)
	}

	return &poolCrawler{
		pool:        rod.NewBrowserPool(size),
		poolSize:    size,
		config:      cfg,
		browserURLs: urls,
		logger:      logger,
	}, nil
}

// newBrowser 从预启动列表中取出一个浏览器并连接
func (c *poolCrawler) newBrowser() (*rod.Browser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.browserURLs) == 0 {
		return nil, fmt.Errorf("没有可用的浏览器实例")
	}
	url := c.browserURLs[0]
	c.browserURLs = c.browserURLs[1:]

	browser := rod.New().ControlURL(url).Trace(c.config.Trace)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %v", err)
	}
	return browser, nil
}

// Close 关闭爬虫，释放所有浏览器资源
func (c *poolCrawler) Close() {
	c.logger.Info("关闭爬虫", "pool_size", c.poolSize)
	c.pool.Cleanup(func(b *rod.Browser) { b.MustClose() })
}

// Crawl 并发执行爬取任务
func (c *poolCrawler) Crawl(ctx context.Context, params []*CrawlerParam) error {
	if len(params) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan *CrawlerParam, len(params))
	for _, p := range params {
		jobs <- p
	}
	close(jobs)

	errs := make(chan error, len(params))
	workers := min(c.poolSize, len(params))

	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			c.worker(ctx, i, jobs, errs)
		})
	}
	wg.Wait()
	close(errs)

	return collectErrors(errs)
}

// worker 工作协程：从任务通道取任务并处理
func (c *poolCrawler) worker(ctx context.Context, id int, jobs <-chan *CrawlerParam, errs chan<- error) {
	defer c.recoverWorker(id)

	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("worker 退出", "worker", id, "reason", "context cancelled")
			return
		case param, ok := <-jobs:
			if !ok {
				return
			}
			c.handle(ctx, id, errs, param)
		}
	}
}

// recoverWorker 恢复 worker 中的 panic
func (c *poolCrawler) recoverWorker(id int) {
	if r := recover(); r != nil {
		c.logger.Error("worker panic", "worker", id, "panic", r)
	}
}

// handle 处理单个爬取参数
func (c *poolCrawler) handle(ctx context.Context, id int, errs chan<- error, param *CrawlerParam) {
	browser, err := c.pool.Get(c.newBrowser)
	if err != nil {
		errs <- fmt.Errorf("[worker %d] 获取浏览器失败: %v", id, err)
		return
	}
	defer func() {
		c.logger.Debug("归还浏览器", "worker", id, "url", param.URL)
		c.pool.Put(browser)
	}()

	page, err := stealth.Page(browser)
	if err != nil {
		errs <- fmt.Errorf("[worker %d] 创建隐身页面失败: %v", id, err)
		return
	}
	defer c.closePage(id, page)

	if err = c.navigate(page, id, param.URL); err != nil {
		errs <- err
		return
	}

	if router := c.setupInterceptors(ctx, browser, param.NetworkConfigs); router != nil {
		defer router.Stop()
	}

	if err = c.runActions(ctx, page, id, param); err != nil {
		errs <- err
	}
}

// closePage 安全关闭页面
func (c *poolCrawler) closePage(id int, page *rod.Page) {
	c.logger.Debug("关闭页面", "worker", id)
	if err := page.Close(); err != nil {
		c.logger.Warn("关闭页面异常", "worker", id, "error", err)
	}
}

// navigate 导航到目标URL并等待页面稳定
func (c *poolCrawler) navigate(page *rod.Page, id int, url string) error {
	c.logger.Info("打开页面", "worker", id, "url", url)

	if err := page.Navigate(url); err != nil {
		return fmt.Errorf("[worker %d] 导航失败: %v", id, err)
	}
	/*
		if err := page.WaitStable(10 * time.Second); err != nil {
			return fmt.Errorf("[worker %d] 页面稳定等待超时: %v", id, err)
		}
	*/
	// 用 WaitLoad 替代 WaitStable，页面加载完即可
	if err := page.WaitLoad(); err != nil {
		return fmt.Errorf("[worker %d] 页面加载等待失败: %v", id, err)
	}
	return nil
}

// setupInterceptors 设置网络拦截器，返回 router 供调用方 Stop
func (c *poolCrawler) setupInterceptors(ctx context.Context, browser *rod.Browser, configs []*NetworkConfig) *rod.HijackRouter {
	if len(configs) == 0 {
		return nil
	}

	router := browser.HijackRequests()
	for _, nc := range configs {
		if nc == nil || nc.URLPattern == "" {
			continue
		}
		nc := nc
		router.MustAdd(nc.URLPattern, func(hijack *rod.Hijack) {
			c.handleHijack(ctx, nc, hijack)
		})
	}

	go router.Run()
	return router
}

// handleHijack 处理单个网络拦截
func (c *poolCrawler) handleHijack(ctx context.Context, nc *NetworkConfig, hijack *rod.Hijack) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if err := hijack.LoadResponse(http.DefaultClient, true); err != nil {
		c.logger.Warn("拦截加载失败", "pattern", nc.URLPattern, "error", err)
		return
	}

	if nc.ProcessFunc == nil {
		return
	}

	if err := nc.ProcessFunc(ctx, &NetworkResponse{
		Url:        hijack.Request.URL().String(),
		UrlPattern: nc.URLPattern,
		Body:       hijack.Response.Body(),
	}); err != nil {
		c.logger.Warn("拦截处理失败", "pattern", nc.URLPattern, "error", err)
	}
}

// runActions 按顺序执行页面操作
func (c *poolCrawler) runActions(ctx context.Context, page *rod.Page, id int, param *CrawlerParam) error {
	if len(param.Actions) == 0 {
		return nil
	}

	includes := extractPatterns(param.NetworkConfigs)

	for i, action := range param.Actions {
		if err := action.Validate(); err != nil {
			return fmt.Errorf("[worker %d] action[%d] 验证失败: %v", id, i, err)
		}

		if err := c.executeAction(ctx, page, action, includes); err != nil {
			return fmt.Errorf("[worker %d] action[%d] %v", id, i, err)
		}
	}
	return nil
}

// executeAction 执行单个操作并等待
func (c *poolCrawler) executeAction(ctx context.Context, page *rod.Page, action Action, waitPatterns []string) error {
	var err error
	var delay time.Duration

	switch a := action.(type) {
	case *ClickAction:
		delay = a.Delay
		err = c.doClick(page, a.Selector)
	case *ClickXAction:
		delay = a.Delay
		err = c.doClickX(page, a.Selector)
	case *ScrollAction:
		delay = a.Delay
		err = c.doScroll(page, a.ScrollY)
	case *JavaScriptAction:
		delay = a.Delay
		err = c.doJavaScript(ctx, page, a)
	default:
		return fmt.Errorf("未知操作类型: %T", a)
	}

	if err != nil {
		return err
	}
	return c.waitIdle(ctx, page, delay, waitPatterns)
}

// doClick CSS选择器点击
func (c *poolCrawler) doClick(page *rod.Page, selector string) error {
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("未找到元素 %q: %v", selector, err)
	}
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// doClickX XPath点击
func (c *poolCrawler) doClickX(page *rod.Page, selector string) error {
	el, err := page.ElementX(selector)
	if err != nil {
		return fmt.Errorf("未找到XPath元素 %q: %v", selector, err)
	}
	return el.Click(proto.InputMouseButtonLeft, 1)
}

// doScroll 页面滚动
func (c *poolCrawler) doScroll(page *rod.Page, scrollY int) error {
	_, err := page.Eval(`(dy) => window.scrollBy({top: dy, behavior: 'smooth'})`, scrollY)
	return err
}

// doJavaScript 执行JS并回调处理结果
func (c *poolCrawler) doJavaScript(ctx context.Context, page *rod.Page, a *JavaScriptAction) error {
	result, err := page.Eval(a.JavaScript, a.JavaScriptArgs...)
	if err != nil {
		return fmt.Errorf("JS执行失败: %v", err)
	}

	if a.ProcessFunc == nil {
		return nil
	}

	data, err := result.Value.MarshalJSON()
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("获取页面信息失败: %v", err)
	}

	return a.ProcessFunc(ctx, &HtmlContent{Url: info.URL, Content: data})
}

// waitIdle 等待网络空闲 + 自定义延迟
func (c *poolCrawler) waitIdle(ctx context.Context, page *rod.Page, delay time.Duration, patterns []string) error {
	if err := c.checkCtx(ctx); err != nil {
		return err
	}

	page.WaitRequestIdle(500*time.Millisecond, patterns, nil, nil)

	if delay > 0 {
		return c.sleep(ctx, delay)
	}
	return nil
}

// checkCtx 检查 context 是否已取消
func (c *poolCrawler) checkCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// sleep 可被 context 取消的等待
func (c *poolCrawler) sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// extractPatterns 提取网络配置中的 URL 匹配模式
func extractPatterns(configs []*NetworkConfig) []string {
	patterns := make([]string, 0, len(configs))
	for _, nc := range configs {
		if nc != nil && nc.URLPattern != "" {
			patterns = append(patterns, nc.URLPattern)
		}
	}
	return patterns
}

// collectErrors 汇总错误通道中的错误，使用 errors.Join 保留错误树
func collectErrors(errs <-chan error) error {
	var list []error
	for e := range errs {
		list = append(list, e)
	}
	return errors.Join(list...)
}
