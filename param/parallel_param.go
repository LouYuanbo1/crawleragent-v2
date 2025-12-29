package param

import "crawleragent-v2/internal/domain/model"

type ParallelNetworkConfig struct {
	URLPattern   string                                      `json:"url_pattern"`
	RespChanSize int                                         `json:"resp_chan_size"`
	ToDocFunc    func(body []byte) ([]model.Document, error) `json:"to_doc_func"`
}

type ParallelCrawlerParam struct {
	URL            string                   `json:"url"`
	NetworkConfigs []*ParallelNetworkConfig `json:"network_configs"`
	Actions        []Action                 `json:"actions"`
}
