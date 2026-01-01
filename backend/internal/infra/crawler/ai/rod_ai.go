package ai

import (
	"context"
	"crawleragent-v2/internal/config"
	"crawleragent-v2/internal/infra/crawler"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

type aiCrawler struct {
	// browser 是全局浏览器实例，用于创建页面和连接
	browser *rod.Browser
	page    *rod.Page
	router  *rod.HijackRouter
}

func InitAICrawler(cfg *config.Config) (AICrawler, error) {
	instanceDataDir := fmt.Sprintf("%s/instance_%d", cfg.Rod.UserDataDir, 0)
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
	)
	urlStr, err := url.Launch()
	if err != nil {
		return nil, fmt.Errorf("启动浏览器失败: %v", err)
	}
	log.Printf("浏览器可以连接的URL: %s", urlStr)
	browser := rod.New().ControlURL(urlStr).Trace(cfg.Rod.Trace)
	err = browser.Connect()
	if err != nil {
		return nil, fmt.Errorf("连接浏览器失败: %v", err)
	}
	page, err := stealth.Page(browser)
	if err != nil {
		return nil, fmt.Errorf("应用Stealth插件失败: %v", err)
	}
	return &aiCrawler{
		browser: browser,
		page:    page,
		router:  nil,
	}, nil
}

func (c *aiCrawler) CloseAll() error {
	// 关闭路由器
	if c.router != nil {
		err := c.router.Stop()
		log.Printf("运行结束,关闭路由器")
		if err != nil {
			log.Printf("关闭路由器失败: %v", err)
			return fmt.Errorf("关闭路由器失败: %v", err)
		}
	}
	// 关闭页面
	err := c.page.Close()
	log.Printf("运行结束,关闭页面")
	if err != nil {
		log.Printf("关闭页面失败: %v", err)
		return fmt.Errorf("关闭页面失败: %v", err)
	}
	// 关闭浏览器
	err = c.browser.Close()
	log.Printf("运行结束,关闭浏览器")
	if err != nil {
		log.Printf("关闭浏览器失败: %v", err)
		return fmt.Errorf("关闭浏览器失败: %v", err)
	}

	return nil
}

func (c *aiCrawler) CloseRouter() error {
	// 关闭路由器
	if c.router != nil {
		err := c.router.Stop()
		log.Printf("关闭路由器")
		if err != nil {
			log.Printf("关闭路由器失败: %v", err)
			return fmt.Errorf("关闭路由器失败: %v", err)
		}
	}
	return nil
}

func (c *aiCrawler) NavigateURL(url string) error {
	err := c.page.Navigate(url)
	if err != nil {
		log.Printf("导航到URL失败: %v", err)
		return fmt.Errorf("导航失败: %v", err)
	}

	c.page.MustWaitStable()

	time.Sleep(2 * time.Second)
	return nil
}

func (c *aiCrawler) GetHTML() (string, error) {

	html, err := c.page.HTML()
	if err != nil {
		log.Printf("获取HTML失败: %v", err)
		return "", fmt.Errorf("获取HTML失败: %v", err)
	}
	return html, nil
}

func preprocessIncludeTags(includeTags, excludeTags []string) ([]string, []string) {
	// 如果任一列表为空，则没有工作可做
	if len(includeTags) == 0 || len(excludeTags) == 0 {
		return includeTags, nil
	}

	// 1. 将excludeTags转换为map，用于快速查找
	excludeSet := make(map[string]struct{})
	for _, tag := range excludeTags {
		excludeSet[tag] = struct{}{}
	}

	// 2. 遍历includeTags，分离出需保留和需移除的标签
	newIncludes := make([]string, 0, len(includeTags))
	overlaps := make([]string, 0)

	for _, tag := range includeTags {
		if _, found := excludeSet[tag]; found {
			// 如果是重合标签，则添加到重叠列表
			overlaps = append(overlaps, tag)
		} else {
			// 否则，添加到新的包含列表
			newIncludes = append(newIncludes, tag)
		}
	}
	return newIncludes, overlaps
}

func (c *aiCrawler) ExecuteActions(actions []param.Action, waitIncludes, waitExcludes []string) error {
	if len(actions) == 0 {
		return nil
	}
	for _, action := range actions {
		err := action.Validate()
		if err != nil {
			return fmt.Errorf("操作验证失败: %v", err)
		}

		switch a := action.(type) {
		case *param.ClickAction:
			element, err := c.page.Element(a.Selector)
			if err != nil {
				return fmt.Errorf("点击操作失败: %v", err)
			}
			err = element.Click(proto.InputMouseButtonLeft, 1)
			if err != nil {
				return fmt.Errorf("点击操作失败: %v", err)
			}
			c.page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
			time.Sleep(a.Delay)
		case *param.ScrollAction:
			_, err = c.page.Eval(
				`
				(scrollY) => {
					window.scrollTo({
						top: scrollY,
						behavior: 'smooth'
					});
				}
			`, a.ScrollY)
			c.page.WaitRequestIdle(500*time.Millisecond, waitIncludes, waitExcludes, nil)
			time.Sleep(a.Delay)
		default:
			return fmt.Errorf("未知操作类型: %T", a)
		}
	}
	return nil
}

func (c *aiCrawler) preProcessHTML(tempPage *rod.Page, candidates []string) (string, error) {
	if candidates == nil {
		candidates = []string{
			"main", "[role=\"main\"]",
			"#content", ".content", ".post-content",
		}
	}

	// 修改：增加移除script标签的功能
	jsResult, err := tempPage.Eval(`
        (candidateSelectors) => {
            function findMainContent(candidateSelectors) {
                // 确保 candidateSelectors 是数组
                if (!Array.isArray(candidateSelectors)) {
                    candidateSelectors = [];
                }
                
                for (const selector of candidateSelectors) {
                    const element = document.querySelector(selector);
                    if (element && isSubstantialElement(element)) {
                        return element;
                    }
                }
                
                return document.body;
            }

            function isSubstantialElement(element) {
                if (!element || !element.textContent) {
                    return false;
                }
                const text = element.textContent || '';
                const trimmedText = text.replace(/\s+/g, '').trim();
                return trimmedText.length > 50; // 至少有 50 个字符
            }
            
            function removeScriptTags(htmlElement) {
                // 克隆节点以避免修改原始页面
                const clonedElement = htmlElement.cloneNode(true);
                
                // 移除所有 script 标签
                const scripts = clonedElement.querySelectorAll('script');
                scripts.forEach(script => script.remove());
                
                // 移除内联事件处理程序
                const allElements = clonedElement.querySelectorAll('*');
                allElements.forEach(el => {
                    // 移除常见的内联事件属性
                    const eventAttrs = ['onclick', 'onmouseover', 'onmouseout', 'onload', 'onerror', 'onchange', 'onsubmit', 'onkeydown', 'onkeyup', 'onkeypress'];
                    eventAttrs.forEach(attr => {
                        if (el.hasAttribute(attr)) {
                            el.removeAttribute(attr);
                        }
                    });
                    
                    // 移除 javascript: 开头的 href
                    if (el.hasAttribute('href') && el.getAttribute('href').trim().toLowerCase().startsWith('javascript:')) {
                        el.removeAttribute('href');
                    }
                });
                
                return clonedElement;
            }
            
            const mainContent = findMainContent(candidateSelectors);
            if (!mainContent) {
                return document.body.outerHTML;
            }
            
            // 移除 script 标签和其他不需要的 JavaScript 相关内容
            const cleanedElement = removeScriptTags(mainContent);
            return cleanedElement.outerHTML;
        }
    `, candidates) // 参数直接传给函数

	if err != nil {
		return "", fmt.Errorf("执行JS失败: %v", err)
	}

	if jsResult == nil {
		return "", fmt.Errorf("未找到主要内容")
	}

	//log.Printf("预处理HTML结果:\n%s", jsResult.Value.String())

	return jsResult.Value.String(), nil
}

func (c *aiCrawler) CleanHTML(html string, candidates, includeTags, excludeTags []string) (string, error) {
	newIncludes, overlapTags := preprocessIncludeTags(includeTags, excludeTags)
	if len(overlapTags) > 0 {
		log.Printf("警告: includeTags 与 excludeTags 存在重叠: %v。将按'排除优先'规则处理。", overlapTags)
	}

	tempPage, err := c.browser.Page(proto.TargetCreateTarget{
		URL: "about:blank",
	})
	if err != nil {
		return "", fmt.Errorf("创建临时页面失败: %v", err)
	}
	defer tempPage.Close()

	// 设置 HTML 内容
	err = tempPage.SetDocumentContent(html)
	if err != nil {
		return "", fmt.Errorf("设置HTML内容失败: %v", err)
	}

	tempPage.MustWaitLoad()

	preprocessedHTML, err := c.preProcessHTML(tempPage, candidates)
	if err != nil {
		log.Printf("预处理HTML失败: %v", err)
		return "", fmt.Errorf("预处理HTML失败: %v", err)
	}

	// 修复：使用结构化的清洗方法
	jsResult, err := tempPage.Eval(`
        (htmlString, includeSelectors, excludeSelectors) => {
            // 创建临时容器
            const container = document.createElement('div');
            container.innerHTML = htmlString;
            
            // 方法1：结构化过滤 - 保留父容器
            function structuredFilter(root) {
                // 1. 先应用排除规则
                if (excludeSelectors && excludeSelectors.length > 0) {
                    excludeSelectors.forEach(selector => {
                        const elements = root.querySelectorAll(selector);
                        elements.forEach(el => {
                            if (el.parentNode) {
                                el.parentNode.removeChild(el);
                            }
                        });
                    });
                }
                
                // 2. 如果没有包含规则，返回整个内容
                if (!includeSelectors || includeSelectors.length === 0) {
                    return root;
                }
                
                // 3. 创建结果容器
                const resultContainer = document.createElement('div');
                
                // 4. 对于每个包含选择器，保留完整的子树
                includeSelectors.forEach(selector => {
                    const elements = root.querySelectorAll(selector);
                    elements.forEach(element => {
                        // 保留元素及其所有内容
                        const clone = element.cloneNode(true);
                        resultContainer.appendChild(clone);
                    });
                });
                
                return resultContainer;
            }
            
            // 执行结构化过滤
            const result = structuredFilter(container);
            return result.innerHTML;
        }
    `, preprocessedHTML, newIncludes, excludeTags)

	if err != nil {
		return "", fmt.Errorf("执行JS失败: %v", err)
	}

	if jsResult == nil {
		return "", fmt.Errorf("未找到主要内容")
	}

	cleanedHTML := jsResult.Value.String()

	log.Printf("清洗HTML结果:\n%s", cleanedHTML)

	return cleanedHTML, nil
}

func (c *aiCrawler) SetListener(ctx context.Context, urlPatterns []string, respCh chan *types.NetworkResponse) {
	if c.router == nil {
		c.router = c.browser.HijackRequests()
	}
	for _, urlPattern := range urlPatterns {
		c.router.MustAdd(urlPattern, func(hijack *rod.Hijack) {
			select {
			case <-ctx.Done():
				return
			default:
			}
			hijack.MustLoadResponse()
			body := hijack.Response.Body()
			log.Printf("监听成功: %s, 响应长度: %d", urlPattern, len(body))
			respCh <- &types.NetworkResponse{
				Url:        hijack.Request.URL().String(),
				UrlPattern: urlPattern,
				Body:       body,
			}
		})
	}
}

func (c *aiCrawler) RouterRun() {
	go c.router.Run()
}
