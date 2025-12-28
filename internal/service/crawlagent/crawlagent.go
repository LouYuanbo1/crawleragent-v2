package crawlagent

import (
	"context"
	"crawleragent-v2/internal/infra/crawler/ai"
	"crawleragent-v2/internal/infra/llm"
	"crawleragent-v2/param"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type CrawlAgent interface {
	Stream(ctx context.Context, url string, params param.AICrawlerParam) error
	Invoke(ctx context.Context, url string, params param.AICrawlerParam) error
}

type crawlAgent struct {
	llm     llm.LLM
	crawler ai.AICrawler
	graph   compose.Runnable[map[string]any, map[string]any]
}

func InitCrawlAgent(ctx context.Context, llm llm.LLM, crawler ai.AICrawler, prompt *prompt.DefaultChatTemplate) (CrawlAgent, error) {
	graph, err := initAgentGraph(ctx, llm, crawler, prompt)
	if err != nil {
		log.Printf("创建流程图失败: %v", err)
		return nil, fmt.Errorf("创建流程图失败: %w", err)
	}
	return &crawlAgent{llm: llm, crawler: crawler, graph: graph}, nil
}

func initAgentGraph(ctx context.Context, llm llm.LLM, crawler ai.AICrawler, prompt *prompt.DefaultChatTemplate) (compose.Runnable[map[string]any, map[string]any], error) {
	graph := compose.NewGraph[map[string]any, map[string]any]()

	err := graph.AddLambdaNode("processHTML", ProcessHTML(crawler))
	if err != nil {
		log.Printf("Error adding lambda node: %v", err)
		return nil, err
	}

	err = graph.AddChatTemplateNode("prompt", prompt)
	if err != nil {
		log.Printf("Error adding prompt template node: %v", err)
		return nil, err
	}

	err = graph.AddChatModelNode("llm", llm.Model(), compose.WithOutputKey("finalResponse"))
	if err != nil {
		log.Printf("Error adding LLM node: %v", err)
		return nil, err
	}

	err = graph.AddEdge(compose.START, "processHTML")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("processHTML", "prompt")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("prompt", "llm")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("llm", compose.END)
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	compiledGraph, _ := graph.Compile(ctx)
	return compiledGraph, nil
}

func (c *crawlAgent) Invoke(ctx context.Context, url string, params param.AICrawlerParam) error {
	result, err := c.graph.Invoke(ctx, map[string]any{
		"url":    url,
		"params": params,
	})
	if err != nil {
		log.Printf("Failed to invoke graph: %v", err)
		return err
	}

	// 从结果中提取最终回复
	if finalResponse, ok := result["finalResponse"].(*schema.Message); ok {
		fmt.Println(finalResponse.Content)
		return nil
	}

	fmt.Println("抱歉，我无法理解您的请求。")
	return nil
}

func (c *crawlAgent) Stream(ctx context.Context, url string, params param.AICrawlerParam) error {
	result, err := c.graph.Stream(ctx, map[string]any{
		"url":    url,
		"params": params,
	})
	if err != nil {
		log.Printf("Failed to invoke graph: %v", err)
		return err
	}

	for {
		chunk, err := result.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Printf("\n\n")
			break
		}
		if err != nil {
			log.Printf("Error receiving chunk: %v", err)
			return err
		}
		if msg, ok := chunk["finalResponse"].(*schema.Message); ok {
			fmt.Print(msg.Content)
		}
	}
	return nil
}
