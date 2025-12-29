package parallel

import (
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"fmt"
)

type ParallelNetworkRuntime struct {
	URLPattern string                `json:"url_pattern"`
	RespChan   chan types.UrlContent `json:"resp_chan"`
}

type ParallelCrawlerRuntime struct {
	URL            string                    `json:"url"` // 保留原有字段
	NetworkConfigs []*ParallelNetworkRuntime `json:"network_config"`
	Actions        []param.Action            `json:"actions"` // 复用原有Action类型
}

type JavaScriptActionRuntime struct {
	param.BaseParams
	JavaScript     string                `json:"javascript"`      // 要执行的 JavaScript 代码
	JavaScriptArgs []any                 `json:"javascript_args"` // JavaScript 参数
	ContentChan    chan types.UrlContent `json:"content_chan"`    // 用于传递 JavaScript 执行结果，不序列化
}

func (j *JavaScriptActionRuntime) Validate() error {
	if j.JavaScript == "" {
		return fmt.Errorf("JavaScript操作必须指定JavaScript代码")
	}
	return nil
}
