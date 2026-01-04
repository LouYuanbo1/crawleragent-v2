package utils

import (
	"crawleragent-v2/internal/config"

	"github.com/gin-gonic/gin"
)

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
