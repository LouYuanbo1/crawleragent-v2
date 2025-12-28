package param

import "crawleragent-v2/internal/domain/model"

type ParallelHTMLConfig struct {
	ContentChanSize int                                                    `json:"content_chan_size"`
	ToDocJsFunc     func(js string, args ...any) ([]model.Document, error) `json:"to_doc_js_func"`
}

type ParallelNetworkConfig struct {
	URLPattern   string                                      `json:"url_pattern"`
	RespChanSize int                                         `json:"resp_chan_size"`
	ToDocFunc    func(body []byte) ([]model.Document, error) `json:"to_doc_func"`
}

type ParallelCrawlerParam struct {
	URL string `json:"url"`
	//HTMLConfigs    []*ParallelHTMLConfig    `json:"html_configs"`
	NetworkConfigs []*ParallelNetworkConfig `json:"network_configs"`
	Actions        []Action                 `json:"actions"`
}
