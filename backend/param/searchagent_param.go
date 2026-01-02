package param

import (
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
)

type SearchConfig struct {
	MaxResults int
	Region     duckduckgo.Region
	Timeout    time.Duration
}

type Agent struct {
	DuckDuckGoSearch SearchConfig
}

type QueryWithPrompt struct {
	Index           string `json:"index"`
	Query           string `json:"query"`
	PromptEsRAGMode string `json:"promptEsRAGMode"`
	PromptChatMode  string `json:"promptChatMode"`
}
