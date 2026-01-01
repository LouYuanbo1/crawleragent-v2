package main

import (
	"context"
	"crawleragent-v2/internal/config"
	"crawleragent-v2/internal/infra/crawler/ai"
	"crawleragent-v2/internal/infra/llm"
	"crawleragent-v2/internal/service/crawlagent"
	"crawleragent-v2/param"
	"log"
	"time"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	crawler, err := ai.InitAICrawler(cfg)
	if err != nil {
		log.Fatalf("初始化浏览器池失败: %v", err)
	}

	ctx := context.Background()

	llm, err := llm.InitLLM(ctx, cfg)
	if err != nil {
		log.Fatalf("初始化LLM失败: %v", err)
	}

	prompt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`角色:你是一位信息提取工具,负责从HTML中提取信息并根据json风格的schema中的定义进行格式化。`),
		schema.SystemMessage(`输入schema:\n{schema}`),
		schema.SystemMessage(`以下是处理后的HTML:\n{cleanedHTML}\n\n搜寻其中的内容并将内容填入schema中,如果HTML中没有相关内容,忽略该字段。`),
		schema.SystemMessage(`以下是监听的Json结果:\n{networkResponses}\n\n搜寻其中的内容并将内容填入schema中,如果Json中没有相关内容,忽略该字段。`),
	)

	agent, err := service.InitCrawlAgentService(ctx, llm, crawler, prompt)
	if err != nil {
		log.Fatalf("初始化CrawlAgent失败: %v", err)
	}

	agent.Invoke(ctx, "https://www.cnblogs.com/", param.AICrawlerParam{
		Formats: param.Formats{
			Schema: param.Schema{
				Type: "array",
				Properties: map[string]param.Schema{
					"summary": {
						Type:        "string",
						Description: "博客的摘要",
					},
				},
			},
		},
		HTMLConfig: &param.AIHTMLConfig{
			OnlyMainContent: true,
			IncludeTags:     []string{"p"},
		},
	})

	agent.Invoke(ctx, "https://www.bilibili.com/", param.AICrawlerParam{
		Formats: param.Formats{
			Schema: param.Schema{
				Type: "array",
				Properties: map[string]param.Schema{
					"title": {
						Type:        "string",
						Description: "视频的标题",
					},
				},
			},
		},
		NetworkConfig: &param.AINetworkConfig{
			URLPatterns:  []string{"https://api.bilibili.com/x/web-interface/index/ogv/rcmd*"},
			RespChanSize: 100,
		},
		Actions: []param.Action{
			&param.ScrollAction{
				BaseParams: param.BaseParams{
					Delay: 1000 * time.Millisecond,
				},
				ScrollY: 1000,
			},
			&param.ScrollAction{
				BaseParams: param.BaseParams{
					Delay: 1000 * time.Millisecond,
				},
				ScrollY: 1500,
			},
			&param.ScrollAction{
				BaseParams: param.BaseParams{
					Delay: 1000 * time.Millisecond,
				},
				ScrollY: 1200,
			},
		},
	})

	crawler.CloseAll()
}
