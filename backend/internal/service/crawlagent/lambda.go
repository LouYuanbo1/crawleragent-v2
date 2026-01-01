package crawlagent

import (
	"context"
	"crawleragent-v2/internal/infra/crawler/ai"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/compose"
)

func ProcessHTML(crawler ai.AICrawler) *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, state map[string]any) (map[string]any, error) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		params, ok := state["params"].(param.AICrawlerParam)
		if !ok {
			return nil, fmt.Errorf("params not found in state")
		}
		url, ok := state["url"].(string)
		if !ok {
			return nil, fmt.Errorf("url not found in state")
		}

		state["networkResponses"] = ""
		state["cleanedHTML"] = ""

		if params.NetworkConfig != nil {
			respChan := make(chan *types.NetworkResponse, params.NetworkConfig.RespChanSize)
			defer close(respChan) // 确保通道被关闭

			// 设置监听器
			crawler.SetListener(ctx, params.NetworkConfig.URLPatterns, respChan)
			crawler.RouterRun()
			defer crawler.CloseRouter() // 确保路由器被关闭

			var builder strings.Builder
			var respCount int

			go func() {
				for resp := range respChan {
					log.Printf("收到网络响应: %v, 匹配模式: %v, 响应长度: %d",
						resp.Url, resp.UrlPattern, len(resp.Body))
					builder.WriteString(fmt.Sprintf("第%d个网络响应: %v\n", respCount+1, resp.Body))
					respCount++
				}
				log.Printf("处理完成,网络响应 %d 个: %s", respCount, builder.String())
			}()
		}

		// 导航和执行操作
		err := crawler.NavigateURL(url)
		if err != nil {
			return nil, fmt.Errorf("crawl html failed: %w", err)
		}

		// 执行操作后立即关闭相关资源
		if params.NetworkConfig != nil {
			err = crawler.ExecuteActions(params.Actions, params.NetworkConfig.URLPatterns, nil)
		} else {
			err = crawler.ExecuteActions(params.Actions, nil, nil)
		}
		if err != nil {
			return nil, fmt.Errorf("execute actions failed: %w", err)
		}

		// 处理HTML
		if params.HTMLConfig != nil {
			html, err := crawler.GetHTML()
			if err != nil {
				return nil, fmt.Errorf("get html failed: %w", err)
			}
			cleanedHTML, err := crawler.CleanHTML(html, params.HTMLConfig.Candidates,
				params.HTMLConfig.IncludeTags, params.HTMLConfig.ExcludeTags)
			if err != nil {
				return nil, fmt.Errorf("clean html failed: %w", err)
			}
			state["cleanedHTML"] = cleanedHTML
		}

		state["schema"] = params.Formats.Schema
		return state, nil
	})
}
