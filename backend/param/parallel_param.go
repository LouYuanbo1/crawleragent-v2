package param

import (
	"context"
	"crawleragent-v2/types"
)

type ParallelNetworkConfig struct {
	URLPattern string `json:"url_pattern"`
	//ToDocFunc   func(ctx context.Context, content types.UrlContent) ([]model.Document, error)
	ProcessFunc func(ctx context.Context, content types.UrlContent) error
}

type ParallelCrawlerParam struct {
	URL            string                   `json:"url"`
	NetworkConfigs []*ParallelNetworkConfig `json:"network_configs"`
	Actions        []Action                 `json:"actions"`
}
