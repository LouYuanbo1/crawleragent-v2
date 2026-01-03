package main

import (
	"context"
	"crawleragent-v2/internal/config"
	docController "crawleragent-v2/internal/controller/document"
	searchAgentController "crawleragent-v2/internal/controller/searchagent"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/llm"
	"crawleragent-v2/internal/infra/persistence/es"
	service "crawleragent-v2/internal/service/searchagent"
	"crawleragent-v2/param"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/gin-gonic/gin"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	ctx := context.Background()

	router := gin.Default()

	fmt.Printf("Chromedp UserDataDir: %s\n", appcfg.Rod.UserDataDir)

	client, err := es.InitTypedEsClient(appcfg, 1)
	if err != nil {
		log.Fatalf("初始化ES客户端失败: %v", err)
	}

	embedder, err := embedding.InitEmbedder(ctx, appcfg, 5, 1)
	if err != nil {
		log.Fatalf("初始化Embedder失败: %v", err)
	}

	llm, err := llm.InitLLM(ctx, appcfg)
	if err != nil {
		log.Fatalf("初始化LLM失败: %v", err)
	}

	paramSearchAgent := param.Agent{
		DuckDuckGoSearch: param.SearchConfig{
			MaxResults: 5,
			Region:     duckduckgo.RegionCN,
			Timeout:    20 * time.Second,
		},
	}

	searchAgent, err := service.InitSearchAgentService(ctx, llm, client, embedder, &paramSearchAgent)
	if err != nil {
		log.Fatalf("初始化SearchAgentService失败: %v", err)
	}

	docController := docController.InitDocumentController(client)
	docController.RegisterRoutes(router)

	searchAgentController := searchAgentController.InitSearchAgentController(searchAgent)
	searchAgentController.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
