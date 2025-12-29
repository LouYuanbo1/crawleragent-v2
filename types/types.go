package types

type UrlContent interface {
	GetUrl() string
	GetUrlPattern() string
	GetBody() []byte
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

func (n *NetworkResponse) GetBody() []byte {
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

func (h *HtmlContent) GetBody() []byte {
	return h.Content
}
