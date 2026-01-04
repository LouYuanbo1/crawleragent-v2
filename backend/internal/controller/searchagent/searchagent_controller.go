package controller

import (
	"crawleragent-v2/internal/config"
	"crawleragent-v2/internal/middleware"
	service "crawleragent-v2/internal/service/searchagent"
	"crawleragent-v2/param"
	utils "crawleragent-v2/utils/gin"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type SearchAgentController struct {
	searchAgentService service.SearchAgentService
}

func InitSearchAgentController(searchAgentService service.SearchAgentService) *SearchAgentController {
	return &SearchAgentController{
		searchAgentService: searchAgentService,
	}
}

func (sc *SearchAgentController) RegisterRoutes(r *gin.Engine, appcfg *config.Config) {
	group := r.Group("/api/searchagent")
	{
		group.POST("/test", sc.TestInvoke)
		group.POST("/setting", middleware.WithConfig(appcfg), sc.Setting)
		group.GET("/setting", middleware.WithConfig(appcfg), sc.GetSetting)
		group.POST("", sc.Invoke)
	}
}

type SearchAgentRequest struct {
	param.QueryWithPrompt
}

func (sc *SearchAgentController) TestInvoke(gctx *gin.Context) {
	var searchAgentReq SearchAgentRequest
	if err := gctx.ShouldBindJSON(&searchAgentReq); err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
		return
	}

	resp, err := sc.searchAgentService.Invoke(gctx.Request.Context(), &searchAgentReq.QueryWithPrompt)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to stream: %s", err.Error()), "data": nil})
		return
	}

	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": resp})
}

func (sc *SearchAgentController) Setting(gctx *gin.Context) {
	cfg, ok := utils.GetConfigFromGinContext(gctx)
	if !ok {
		gctx.JSON(500, gin.H{"code": 500, "msg": "failed to get appcfg", "data": nil})
		return
	}

	var searchAgentReq SearchAgentRequest
	if err := gctx.ShouldBindJSON(&searchAgentReq); err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
		return
	}
	jsonSearchAgentReq, err := json.Marshal(searchAgentReq)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to marshal request: %s", err.Error()), "data": nil})
		return
	}

	timestamp := time.Now().Format("20060102150405")
	err = os.MkdirAll(cfg.Prompt.PromptDir, 0755)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to mkdir: %s", err.Error()), "data": nil})
		return
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s_%s.json", cfg.Prompt.PromptDir, searchAgentReq.Index, timestamp), jsonSearchAgentReq, 0644)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to write file: %s", err.Error()), "data": nil})
		return
	}
	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": "保存文件成功"})
}

func (sc *SearchAgentController) GetSetting(gctx *gin.Context) {
	cfg, ok := utils.GetConfigFromGinContext(gctx)
	if !ok {
		gctx.JSON(500, gin.H{"code": 500, "msg": "failed to get appcfg", "data": nil})
		return
	}

	index := gctx.Query("index")
	if index == "" {
		gctx.JSON(500, gin.H{"code": 500, "msg": "index is required", "data": nil})
		return
	}

	pathSetting := filepath.Join(cfg.Prompt.PromptDir, fmt.Sprintf("%s_*.json", index))
	files, err := filepath.Glob(pathSetting)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to glob: %s", err.Error()), "data": nil})
		return
	}
	if len(files) == 0 {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("no setting file found for index: %s", index), "data": nil})
		return
	}
	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": files})
}

type InvokeRequest struct {
	Query   string `json:"query"`
	Setting string `json:"setting"`
}

func (sc *SearchAgentController) Invoke(gctx *gin.Context) {

	var invokeReq InvokeRequest
	if err := gctx.ShouldBindJSON(&invokeReq); err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
		return
	}

	log.Printf("query: %s, setting: %s", invokeReq.Query, invokeReq.Setting)

	setting, err := os.ReadFile(invokeReq.Setting)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to read file: %s", err.Error()), "data": nil})
		return
	}

	var queryWithPrompt param.QueryWithPrompt
	if err := json.Unmarshal(setting, &queryWithPrompt); err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
		return
	}

	resp, err := sc.searchAgentService.Invoke(gctx.Request.Context(), &param.QueryWithPrompt{
		Index:           queryWithPrompt.Index,
		PromptEsRAGMode: queryWithPrompt.PromptEsRAGMode,
		PromptChatMode:  queryWithPrompt.PromptChatMode,
		Query:           invokeReq.Query,
	})
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to stream: %s", err.Error()), "data": nil})
		return
	}

	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": resp})
}
