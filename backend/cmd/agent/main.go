package main

import (
	"context"
	"crawleragent-v2/config"
	"crawleragent-v2/internal/agent"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	elasticsearchRetriever "github.com/LouYuanbo1/go-eino-agent/tools/retriever/elasticsearch"
	embeddingOllama "github.com/cloudwego/eino-ext/components/embedding/ollama"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/elastic/go-elasticsearch/v9"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	fmt.Printf("Chromedp UserDataDir: %s\n", appcfg.Rod.UserDataDir)

	crawler, err := agent.InitBrowserCrawler(&appcfg.Rod)
	if err != nil {
		log.Fatalf("初始化浏览器爬虫失败: %v", err)
	}
	defer crawler.Close()

	ctx := context.Background()

	chatModel, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
		APIKey: appcfg.Deepseek.APIKey,
		Model:  "deepseek-v4-pro",
	})
	if err != nil {
		fmt.Printf("Error creating chat model: %v", err)
		return
	}

	elasticsearchClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Username:  appcfg.Elasticsearch.Username,
		Password:  appcfg.Elasticsearch.Password,
		Addresses: []string{"http://localhost:9200"},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			// 跳过TLS验证（仅在开发环境中使用）
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		fmt.Printf("Error creating elasticsearch client: %v", err)
		return
	}
	embeddingModel, err := embeddingOllama.NewEmbedder(ctx, &embeddingOllama.EmbeddingConfig{
		Model:   "nomic-embed-text",
		BaseURL: "http://localhost:11434",
	})
	if err != nil {
		fmt.Printf("Error creating embedder: %v", err)
		return
	}
	typedRetriever := elasticsearchRetriever.NewElasticsearchRetriever(elasticsearchClient, embeddingModel, &elasticsearchRetriever.ElasticsearchRetrieverConfig{
		K:               5,
		IndexName:       "boss_jobs",
		VectorFieldName: "embedding",
		NumCandidates:   100,
	})

	agent, err := agent.NewSupervisorAgent(ctx, chatModel, crawler, typedRetriever)
	if err != nil {
		fmt.Printf("Error creating supervisor agent: %v", err)
		return
	}
	/*
		agent.OutputMessage(ctx, "打开CSDN,向下滑动五次，之后获取主页前三个文章标题", func(s string) {
			fmt.Print(s)
		})
	*/
	agent.OutputMessage(ctx, "帮我使用本地搜索寻找一下最近的岗位信息", func(s string) {
		fmt.Print(s)
	})

}
