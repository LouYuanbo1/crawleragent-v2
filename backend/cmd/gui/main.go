package main

import (
	"crawleragent-v2/internal/config"
	controller "crawleragent-v2/internal/controller/document"
	"crawleragent-v2/internal/infra/persistence/es"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	router := gin.Default()

	fmt.Printf("Chromedp UserDataDir: %s\n", appcfg.Rod.UserDataDir)
	client, err := es.InitTypedEsClient(appcfg, 1)
	if err != nil {
		log.Fatalf("初始化ES客户端失败: %v", err)
	}
	docController := controller.InitDocumentController(client)
	docController.RegisterRoutes(router)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
