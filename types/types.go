package types

type NetworkResponse struct {
	Url        string
	UrlPattern string
	Body       string
}

type HtmlContent struct {
	Url             string
	ContentSelector string
	Content         []string
}
