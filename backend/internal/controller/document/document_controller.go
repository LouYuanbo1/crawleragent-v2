package controller

import (
	"crawleragent-v2/internal/infra/persistence/es"
	"fmt"

	"github.com/gin-gonic/gin"
)

type DocumentController struct {
	typedClient es.TypedEsClient
}

func InitDocumentController(typedClient es.TypedEsClient) *DocumentController {
	return &DocumentController{
		typedClient: typedClient,
	}
}

func (dc *DocumentController) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/documents")
	{
		group.GET("/:index", dc.GetDocsByPages)
		group.GET("/indices", dc.GetMapIndexCount)
	}
}

type GetDocsByPagesReq struct {
	Page int `form:"page" binding:"required"`
	Size int `form:"size" binding:"required"`
}

func (dc *DocumentController) GetMapIndexCount(gctx *gin.Context) {
	ctx := gctx.Request.Context()
	mapIndexCount, err := dc.typedClient.GetMapIndexCount(ctx)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to get map index count: %s", err.Error()), "data": nil})
		return
	}
	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": mapIndexCount})
}

func (dc *DocumentController) GetDocsByPages(gctx *gin.Context) {
	index := gctx.Param("index")
	var req GetDocsByPagesReq
	if err := gctx.ShouldBind(&req); err != nil {
		gctx.JSON(400, gin.H{"code": 400, "msg": fmt.Sprintf("invalid request: %s", err.Error()), "data": nil})
		return
	}
	ctx := gctx.Request.Context()
	docs, err := dc.typedClient.GetDocsByPages(ctx, index, req.Page, req.Size)
	if err != nil {
		gctx.JSON(500, gin.H{"code": 500, "msg": fmt.Sprintf("failed to get docs by pages: %s", err.Error()), "data": nil})
		return
	}
	gctx.JSON(200, gin.H{"code": 200, "msg": "success", "data": docs})
}
