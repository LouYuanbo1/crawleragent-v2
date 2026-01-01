package service

import (
	"context"
	"crawleragent-v2/param"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// IntentDetection 意图检测节点,用于识别用户查询的意图,当用户输入以查询模式或搜索模式开头时,将意图设置为"retriever",
// 否则将意图设置为"chatModePrompt"
func IntentDetection() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, ok := state["query"].(string)
		if !ok {
			return nil, errors.New("query not found in state")
		}
		isEsRAGMode := strings.HasPrefix(query, "查询模式")
		if isEsRAGMode {
			state["isEsRAGMode"] = true
		} else {
			state["isEsRAGMode"] = false
		}
		println("isEsRAGMode: ", state["isEsRAGMode"].(bool))
		return state, nil
	})
}

// BranchCondition 分支条件节点,根据用户查询意图,选择下一个节点,当用户输入以查询模式或搜索模式开头时,将选择"retriever"节点,
// 否则将选择"chatModePrompt"节点
func BranchCondition(ctx context.Context, state map[string]any) (string, error) {
	isEsRAGMode, ok := state["isEsRAGMode"].(bool)
	if !ok {
		return "", errors.New("isEsRAGMode not found in state")
	}
	if isEsRAGMode {
		return "retriever", nil
	}
	return "duckDuckGoSearch", nil
}

// Retriever 检索节点,用于根据用户查询意图,从索引中检索相关文档
func Retriever() *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, ok := state["query"].(string)
		if !ok {
			return nil, errors.New("query not found in state")
		}
		fmt.Printf("query: %s", query)
		var embeddings [][]float32
		var err error
		err = compose.ProcessState(ctx, func(ctx context.Context, s *State) error {
			log.Printf("query: %s", query)
			embeddings, err = s.Embedder.Embed(ctx, []string{query})
			if err != nil {
				return err
			}

			if len(embeddings) == 0 || len(embeddings[0]) == 0 {
				return errors.New("嵌入结果为空，无法进行向量检索")
			}

			embedding := embeddings[0]

			docsStr, err := s.TypedEsClient.SearchStrDocsByVector(ctx, s.Doc, embedding, 5, 100)
			if err != nil {
				return err
			}

			state["referenceDocs"] = docsStr
			return nil
		})
		if err != nil {
			return nil, err
		}

		return state, nil
	})
}

func DuckDuckGoSearch(tool tool.InvokableTool, param *param.SearchConfig) *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, state map[string]any) (map[string]any, error) {
		query, ok := state["query"].(string)
		if !ok {
			return nil, errors.New("query not found in state")
		}
		searchReq := &duckduckgo.TextSearchRequest{
			Query: query,
		}
		jsonReq, err := json.Marshal(searchReq)
		if err != nil {
			log.Fatalf("Marshal of search request failed, err=%v", err)
		}
		results := make([]*duckduckgo.TextSearchResult, 0, param.MaxResults)
		resp, err := tool.InvokableRun(ctx, string(jsonReq))
		if err != nil {
			log.Printf("Invoke of duckduckgo failed, err=%v", err)
			return nil, fmt.Errorf("Invoke of duckduckgo failed, err=%w", err)
		}
		var searchResp duckduckgo.TextSearchResponse
		if err = json.Unmarshal([]byte(resp), &searchResp); err != nil {
			log.Fatalf("Unmarshal of search response failed, err=%v", err)
		}

		results = append(results, searchResp.Results...)
		var Builder strings.Builder
		Builder.WriteString("参考文档(JSON格式):\n\n")
		for i, result := range results {
			Builder.WriteString(fmt.Sprintf("\n%d. Title: %s\n", i+1, result.Title))
			Builder.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
			Builder.WriteString(fmt.Sprintf("   Summary: %s\n", result.Summary))
		}
		state["duckDuckGoResults"] = Builder.String()
		return state, nil
	})
}
