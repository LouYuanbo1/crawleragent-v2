package parallel

import (
	"context"
	"crawleragent-v2/internal/config"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"crawleragent-v2/internal/infra/crawler"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

type browserPoolCrawler struct {
	browserPool   rod.Pool[rod.Browser]
	createBrowser func() (*rod.Browser, error)
	controlURLCh  chan string
}

func InitBrowserPoolCrawler(cfg *config.Config, browserPoolSize int) (ParallelCrawler, error) {
	controlURLCh := make(chan string, browserPoolSize)
	for instanceID := range browserPoolSize {

		instanceDataDir := fmt.Sprintf("%s/instance_%d", cfg.Rod.UserDataDir, instanceID)
		err := os.MkdirAll(instanceDataDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("创建实例数据目录失败: %v", err)
		}

		url := crawler.CreateLauncher(cfg.Rod.UserMode,
			crawler.WithBin(cfg.Rod.Bin),
			crawler.WithUserDataDir(instanceDataDir),
			crawler.WithHeadless(cfg.Rod.Headless),
			crawler.WithDisableBlinkFeatures(cfg.Rod.DisableBlinkFeatures),
			crawler.WithIncognito(cfg.Rod.Incognito),
			crawler.WithDisableDevShmUsage(cfg.Rod.DisableDevShmUsage),
			crawler.WithNoSandbox(cfg.Rod.NoSandbox),
			crawler.WithUserAgent(cfg.Rod.UserAgent),
			crawler.WithLeakless(cfg.Rod.Leakless),
			crawler.WithDisableBackgroundNetworking(cfg.Rod.DisableBackgroundNetworking),
			crawler.WithDisableBackgroundTimerThrottling(cfg.Rod.DisableBackgroundTimerThrottling),
			crawler.WithRemoteDebuggingPort(cfg.Rod.BasicRemoteDebuggingPort+instanceID),
		)
		urlStr, err := url.Launch()
		if err != nil {
			return nil, fmt.Errorf("启动浏览器失败: %v", err)
		}

		log.Printf("浏览器可以连接的URL: %s", urlStr)
		controlURLCh <- urlStr
	}
	close(controlURLCh)
	// 创建页面池
	BrowserPool := rod.NewBrowserPool(browserPoolSize)

	createBrowser := func() (*rod.Browser, error) {
		// 从 controlURLCh 中获取 URL
		urlStr := <-controlURLCh
		browser := rod.
			New().
			ControlURL(urlStr).
			Trace(cfg.Rod.Trace) // 开启 CDP 通信追踪（日志会输出请求/响应）
		if err := browser.Connect(); err != nil {
			return nil, fmt.Errorf("连接浏览器失败: %v", err)
		}
		return browser, nil
	}

	return &browserPoolCrawler{
		browserPool:   BrowserPool,
		createBrowser: createBrowser,
		controlURLCh:  controlURLCh,
	}, nil
}

func (c *browserPoolCrawler) Close() {
	log.Printf("开始关闭，停止接收新请求...")

	// 2. 等待一段时间让正在进行的请求完成
	time.Sleep(3 * time.Second) // 可以根据实际情况调整

	log.Printf("关闭 %d 个浏览器连接", len(c.browserPool))
	c.browserPool.Cleanup(func(b *rod.Browser) { b.MustClose() })
}

func (c *browserPoolCrawler) Crawl(ctx context.Context, runtimes []*ParallelCrawlerRuntime) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	runtimeCh := make(chan *ParallelCrawlerRuntime, len(runtimes))
	for _, op := range runtimes {
		runtimeCh <- op
	}
	close(runtimeCh)

	errCh := make(chan error, max(len(runtimeCh), len(c.browserPool)))

	wg := sync.WaitGroup{}
	for i := range min(len(c.browserPool), len(runtimes)) {
		wg.Add(1)
		go func(ctx context.Context, workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done(): // 主动监听 ctx 取消
					log.Printf("worker %d 取消执行，退出", workerID)
					return
				case runtime, ok := <-runtimeCh: // 读取任务
					if !ok { // 通道关闭则退出
						return
					}
					c.processParam(ctx, workerID, errCh, runtime)
				}
			}
		}(ctx, i)
	}
	wg.Wait()

	close(errCh)
	// 收集错误
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d errors occurred: %v", len(errs), errs)
	}
	return nil
}

func (c *browserPoolCrawler) processParam(ctx context.Context, workerID int, errCh chan<- error, runtime *ParallelCrawlerRuntime) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	browser, err := c.browserPool.Get(c.createBrowser)
	if err != nil {
		errCh <- fmt.Errorf("获取浏览器失败: %v", err)
		return
	}
	defer func() {
		log.Printf("将 browser %d 返回池，处理的URL模式: %s 等...", workerID, runtime.NetworkConfigs[0].URLPattern)
		c.browserPool.Put(browser)
	}()

	page, err := stealth.Page(browser)
	if err != nil {
		errCh <- fmt.Errorf("获取页面失败: %v", err)
		return
	}
	defer func() {
		log.Printf("Worker %d 页面关闭", workerID)
		page.MustClose()
	}()

	// 设置所有网络监听器
	if runtime.NetworkConfigs != nil {
		router := c.setListener(ctx, browser, runtime.NetworkConfigs)
		go func() {
			router.Run()
			log.Printf("Worker %d 路由器停止运行", workerID)
		}()
		defer func() {
			log.Printf("Worker %d 路由器关闭", workerID)
			router.Stop()
		}()
	}

	err = c.navigateURL(page, workerID, runtime.URL)
	if err != nil {
		errCh <- fmt.Errorf("处理URL失败: %v", err)
		return
	}

	var waitIncludes []string
	for _, networkConfig := range runtime.NetworkConfigs {
		waitIncludes = append(waitIncludes, networkConfig.URLPattern)
	}

	for _, action := range runtime.Actions {
		err := c.executeAction(page, action, waitIncludes, nil)
		if err != nil {
			errCh <- fmt.Errorf("执行操作失败: %v", err)
			return
		}
	}
}

func (c *browserPoolCrawler) navigateURL(page *rod.Page, workerID int, url string) error {
	// 导航到指定URL
	fmt.Printf("Worker %d 处理: %s\n", workerID, url)

	err := page.Navigate(url)
	if err != nil {
		return fmt.Errorf("导航失败: %v", err)
	}

	page.MustWaitStable()
	time.Sleep(2 * time.Second)

	return nil
}

func (c *browserPoolCrawler) executeAction(page *rod.Page, action param.Action, waitIncludes, waitExcludes []string) error {
	err := action.Validate()
	if err != nil {
		return fmt.Errorf("操作验证失败: %v", err)
	}
	switch a := action.(type) {
	case *param.ClickAction:
		element, err := page.Element(a.Selector)
		if err != nil {
			return fmt.Errorf("点击操作失败: %v", err)
		}
		err = element.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			return fmt.Errorf("点击操作失败: %v", err)
		}
		page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
		time.Sleep(a.Delay)
	case *param.ClickXAction:
		element, err := page.ElementX(a.Selector)
		if err != nil {
			return fmt.Errorf("点击X操作失败: %v", err)
		}
		err = element.Click(proto.InputMouseButtonLeft, 1)
		if err != nil {
			return fmt.Errorf("点击X操作失败: %v", err)
		}
		page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
		time.Sleep(a.Delay)
	case *param.ScrollAction:
		_, err = page.Eval(
			`
					(scrollY) => {
						window.scrollBy({
							top: scrollY,
							behavior: 'smooth'
						});
					}
				`, a.ScrollY)
		page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
		time.Sleep(a.Delay)
	case *JavaScriptActionRuntime:
		// 执行JavaScript操作
		err = c.executeJavaScript(page, a)
		if err != nil {
			return fmt.Errorf("执行JavaScript操作失败: %v", err)
		}
		page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
		time.Sleep(a.Delay)
	default:
		return fmt.Errorf("未知操作类型: %T", a)
	}
	return nil
}

func (c *browserPoolCrawler) setListener(ctx context.Context, browser *rod.Browser, networkConfigs []*ParallelNetworkRuntime) *rod.HijackRouter {
	router := browser.HijackRequests()
	for _, networkConfig := range networkConfigs {
		router.MustAdd(networkConfig.URLPattern, func(hijack *rod.Hijack) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			err := hijack.LoadResponse(http.DefaultClient, true)
			if err != nil {
				log.Printf("加载响应失败: %v", err)
				return
			}
			body := hijack.Response.Body()
			networkConfig.RespChan <- &types.NetworkResponse{
				Url:        hijack.Request.URL().String(),
				UrlPattern: networkConfig.URLPattern,
				Body:       body,
			}
		})
	}
	return router
}

func (c *browserPoolCrawler) executeJavaScript(page *rod.Page, runtime *JavaScriptActionRuntime) error {
	result, err := page.Eval(runtime.JavaScript, runtime.JavaScriptArgs...)
	if err != nil {
		return fmt.Errorf("执行JavaScript失败: %v", err)
	}
	//log.Printf("执行JavaScript: %s", javascript)
	jsonResult, err := result.Value.MarshalJSON()
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	info, err := page.Info()
	if err != nil {
		return fmt.Errorf("获取页面信息失败: %v", err)
	}
	log.Printf("执行JavaScript成功: %d", len(jsonResult))
	runtime.ContentChan <- &types.HtmlContent{
		Url:     info.URL,
		Content: jsonResult,
	}
	return nil
}
