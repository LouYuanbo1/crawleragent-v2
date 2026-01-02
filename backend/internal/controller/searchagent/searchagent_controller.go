package controller

import (
	service "crawleragent-v2/internal/service/searchagent"
	"crawleragent-v2/param"

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

func (sc *SearchAgentController) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/searchagent")
	{
		group.POST("/:index")
	}
}

type StreamRequest struct {
	param.QueryWithPrompt
}

func (sc *SearchAgentController) Stream(gctx *gin.Context) {
	/*
		index := gctx.Param("index")
		var streamReq StreamRequest
		if err := gctx.ShouldBindJSON(&streamReq); err != nil {
			gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
			return
		}

			streamResp, err := sc.searchAgentService.Stream(gctx.Request.Context(), index, streamReq.QueryWithPrompt)
			if err != nil {
				gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to stream: %s", err.Error()), "data": nil})
				return
			}

		gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": streamResp})
	*/
}
