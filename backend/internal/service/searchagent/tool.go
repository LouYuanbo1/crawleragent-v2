package service

import (
	"context"
	"crawleragent-v2/param"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

func InitDuckDuckGo(ctx context.Context, param *param.Agent) (tool.InvokableTool, error) {
	config := &duckduckgo.Config{
		MaxResults: param.DuckDuckGoSearch.MaxResults, // Limit to return 20 results
		Region:     param.DuckDuckGoSearch.Region,
		Timeout:    param.DuckDuckGoSearch.Timeout,
	}
	tool, err := duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		log.Printf("NewTextSearchTool of duckduckgo failed, err=%v", err)
		return nil, fmt.Errorf("NewTextSearchTool of duckduckgo failed, err=%w", err)
	}
	return tool, nil
}
