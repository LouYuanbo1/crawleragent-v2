package crawler

import "context"

// NetworkConfig 网络拦截配置
type NetworkConfig struct {
	URLPattern  string `json:"url_pattern"`
	ProcessFunc func(ctx context.Context, content UrlContent) error
}

// CrawlerParam 单次爬取任务的参数
type CrawlerParam struct {
	URL            string           `json:"url"`
	NetworkConfigs []*NetworkConfig `json:"network_configs"`
	Actions        []Action         `json:"actions"`
}

// UrlContent 爬取内容的统一接口
type UrlContent interface {
	GetUrl() string
	GetUrlPattern() string
	GetContent() []byte
}

// NetworkResponse 网络拦截响应内容
type NetworkResponse struct {
	Url        string
	UrlPattern string
	Body       string
}

func (n *NetworkResponse) GetUrl() string        { return n.Url }
func (n *NetworkResponse) GetUrlPattern() string { return n.UrlPattern }
func (n *NetworkResponse) GetContent() []byte    { return []byte(n.Body) }

// HtmlContent JS执行后获取的页面内容
type HtmlContent struct {
	Url     string
	Content []byte
}

func (h *HtmlContent) GetUrl() string        { return h.Url }
func (h *HtmlContent) GetUrlPattern() string { return "" }
func (h *HtmlContent) GetContent() []byte    { return h.Content }
