package searchagent

import (
	"context"
	"crawleragent-v2/internal/domain/model"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/llm"
	"crawleragent-v2/internal/infra/persistence/es"
	"crawleragent-v2/param"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type State struct {
	Doc           model.Document
	TypedEsClient es.TypedEsClient
	Embedder      embedding.Embedder
}

type SearchAgent interface {
	Stream(ctx context.Context, query string) error
	Invoke(ctx context.Context, query string) error
}

type searchAgent struct {
	llm      llm.LLM
	es       es.TypedEsClient
	embedder embedding.Embedder
	graph    compose.Runnable[map[string]any, map[string]any]
}

func InitSearchAgent(
	ctx context.Context,
	llm llm.LLM,
	es es.TypedEsClient,
	embedder embedding.Embedder,
	param *param.Agent,
) (SearchAgent, error) {
	graph, err := initAgentGraph(ctx, llm, es, embedder, param)
	if err != nil {
		return nil, fmt.Errorf("创建流程图失败: %w", err)
	}
	return &searchAgent{llm: llm, es: es, embedder: embedder, graph: graph}, nil
}

// InitAgent 初始化AgentClient,根据options配置模型和节点
func initAgentGraph(
	ctx context.Context,
	llm llm.LLM,
	typedEsClient es.TypedEsClient,
	embedder embedding.Embedder,
	param *param.Agent,
) (compose.Runnable[map[string]any, map[string]any], error) {
	// 生成State,包含索引名称, TypedEsClient, Embedder 等状态信息
	genState := func(ctx context.Context) *State {
		return &State{
			Doc:           param.Doc,
			TypedEsClient: typedEsClient,
			Embedder:      embedder,
		}
	}

	fmt.Printf("genState: %+v\n", genState(ctx))

	duckDuckGoTool, err := InitDuckDuckGo(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("初始化DuckDuckGo工具失败: %w", err)
	}

	// 初始化Compose图,设置全局状态生成函数
	graph := compose.NewGraph[map[string]any, map[string]any](compose.WithGenLocalState(genState))
	// 添加意图检测节点,用于识别用户查询的意图,当用户输入以查询模式或搜索模式开头时,将意图设置为"retriever",
	// 使用爬取的信息做RAG增强
	err = graph.AddLambdaNode("intentDetection", IntentDetection())
	if err != nil {
		log.Printf("Error adding lambda node: %v", err)
		return nil, err
	}
	// 添加检索节点,用于根据用户查询意图,从索引中检索相关文档
	err = graph.AddLambdaNode("retriever", Retriever())
	if err != nil {
		log.Printf("Error adding lambda node: %v", err)
		return nil, err
	}
	// 添加搜索模式提示节点,用于根据用户查询意图,生成搜索模式的提示
	err = graph.AddChatTemplateNode("searchModePrompt", param.Prompt["EsRAGMode"])
	if err != nil {
		log.Printf("Error adding prompt template node: %v", err)
		return nil, err
	}

	err = graph.AddLambdaNode("duckDuckGoSearch", DuckDuckGoSearch(duckDuckGoTool, &param.DuckDuckGoSearch))
	if err != nil {
		log.Printf("Error adding lambda node: %v", err)
		return nil, err
	}

	// 添加聊天模式提示节点,用于根据用户查询意图,生成聊天模式的提示
	err = graph.AddChatTemplateNode("chatModePrompt", param.Prompt["ChatMode"])
	if err != nil {
		log.Printf("Error adding prompt template node: %v", err)
		return nil, err
	}

	err = graph.AddChatModelNode("llm", llm.Model(), compose.WithOutputKey("finalResponse"))
	if err != nil {
		log.Printf("Error adding LLM node: %v", err)
		return nil, err
	}

	err = graph.AddEdge(compose.START, "intentDetection")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddBranch("intentDetection", compose.NewGraphBranch(BranchCondition, map[string]bool{
		"retriever":        true,
		"duckDuckGoSearch": true,
	}))
	if err != nil {
		log.Printf("Error adding branch: %v", err)
		return nil, err
	}

	err = graph.AddEdge("retriever", "searchModePrompt")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("searchModePrompt", "llm")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("duckDuckGoSearch", "chatModePrompt")
	if err != nil {
		log.Printf("Error adding edge: %v", err)
		return nil, err
	}

	err = graph.AddEdge("chatModePrompt", "llm")
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

func (sa *searchAgent) Invoke(ctx context.Context, query string) error {
	result, err := sa.graph.Invoke(ctx, map[string]any{
		"query": query,
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

func (sa *searchAgent) Stream(ctx context.Context, query string) error {
	result, err := sa.graph.Stream(ctx, map[string]any{
		"query": query,
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
