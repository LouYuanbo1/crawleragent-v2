package param

import (
	"crawleragent-v2/internal/domain/model"
	"time"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/prompt"
)

type PromptType string

const (
	PromptEsRAGMode PromptType = "EsRAGMode"
	PromptChatMode  PromptType = "ChatMode"
)

type SearchConfig struct {
	MaxResults int
	Region     duckduckgo.Region
	Timeout    time.Duration
}

type Agent struct {
	Doc              model.Document
	Prompt           map[PromptType]*prompt.DefaultChatTemplate
	DuckDuckGoSearch SearchConfig
}
