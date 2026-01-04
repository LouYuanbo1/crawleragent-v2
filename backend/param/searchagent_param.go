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
	Index           string `json:"index,omitempty"`
	Query           string `json:"query,omitempty"`
	PromptEsRAGMode string `json:"promptEsRAGMode,omitempty"`
	PromptChatMode  string `json:"promptChatMode,omitempty"`
}
