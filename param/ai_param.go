package param

type Schema struct {
	Type        string            `json:"type"`
	Properties  map[string]Schema `json:"properties"`
	Description string            `json:"description"`
}

type Formats struct {
	Type   string
	Schema Schema
}

type AIHTMLConfig struct {
	OnlyMainContent bool     `json:"only_main_content"`
	Candidates      []string `json:"candidates"`
	IncludeTags     []string `json:"include_tags"`
	ExcludeTags     []string `json:"exclude_tags"`
}

type AINetworkConfig struct {
	URLPatterns  []string `json:"url_patterns"`
	RespChanSize int      `json:"resp_chan_size"`
}

type AICrawlerParam struct {
	Formats       Formats          `json:"formats"`
	HTMLConfig    *AIHTMLConfig    `json:"html_config"`
	NetworkConfig *AINetworkConfig `json:"network_config"`
	Actions       []Action         `json:"actions"`
}
