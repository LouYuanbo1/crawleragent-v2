package parallel

import (
	"crawleragent-v2/param"
	"crawleragent-v2/types"
)

type ParallelHTMLRuntime struct {
	Selectors   []string                `json:"selectors"`
	ContentChan chan *types.HtmlContent `json:"content_chan"`
}

type ParallelNetworkRuntime struct {
	URLPattern string                      `json:"url_pattern"`
	RespChan   chan *types.NetworkResponse `json:"resp_chan"`
}

type ParallelCrawlerRuntime struct {
	URL string `json:"url"` // 保留原有字段
	//HTMLConfigs    []*ParallelHTMLRuntime    `json:"html_config"`
	NetworkConfigs []*ParallelNetworkRuntime `json:"network_config"`
	Actions        []param.Action            `json:"actions"` // 复用原有Action类型
}
