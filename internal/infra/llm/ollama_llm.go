package llm

import (
	"context"
	"crawleragent-v2/internal/config"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/ollama"
)

// 实现AgentClient接口
type llm struct {
	model *ollama.ChatModel
}

func InitLLM(ctx context.Context, config *config.Config) (LLM, error) {
	model, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
		BaseURL: fmt.Sprintf("%s:%d", config.LLM.Host, config.LLM.Port),
		Model:   config.LLM.Model,
	})
	if err != nil {
		log.Printf("Error adding LLM node: %v", err)
		return nil, err
	}
	return &llm{model: model}, nil
}

func (a *llm) Model() *ollama.ChatModel {
	return a.model
}
