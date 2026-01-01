package types

type UrlContent interface {
	GetUrl() string
	GetUrlPattern() string
	GetContent() []byte
}

type NetworkResponse struct {
	Url        string
	UrlPattern string
	Body       string
}

func (n *NetworkResponse) GetUrl() string {
	return n.Url
}
func (n *NetworkResponse) GetUrlPattern() string {
	return n.UrlPattern
}

func (n *NetworkResponse) GetContent() []byte {
	return []byte(n.Body)
}

type HtmlContent struct {
	Url     string
	Content []byte
}

func (h *HtmlContent) GetUrl() string {
	return h.Url
}
func (h *HtmlContent) GetUrlPattern() string {
	return ""
}

func (h *HtmlContent) GetContent() []byte {
	return h.Content
}
