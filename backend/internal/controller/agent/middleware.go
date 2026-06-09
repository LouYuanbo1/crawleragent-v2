package controller

import (
	"crawleragent-v2/config"

	"github.com/gin-gonic/gin"
)

func WithConfig(appcfg *config.Config) gin.HandlerFunc {
	return func(gctx *gin.Context) {
		gctx.Set("appcfg", appcfg)
		gctx.Next()
	}
}


func GetConfigFromGinContext(gctx *gin.Context) (*config.Config, bool) {
	cfg, ok := gctx.Get("appcfg")
	if !ok {
		return nil, false
	}
	appcfg, ok := cfg.(*config.Config)
	if !ok {
		return nil, false
	}
	return appcfg, true
}
