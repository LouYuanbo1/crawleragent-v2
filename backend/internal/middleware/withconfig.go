package middleware

import (
	"crawleragent-v2/internal/config"

	"github.com/gin-gonic/gin"
)

func WithConfig(appcfg *config.Config) gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Set("appcfg", appcfg)
		gctx.Next()
	}
}
