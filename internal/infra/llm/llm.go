package llm

import "github.com/cloudwego/eino-ext/components/model/ollama"

// 定义Agent接口,用于调用模型,并处理模型返回的结果
// 提供Invoke方法返回整个模型的输出结果和Stream方法流式输出
type LLM interface {
	Model() *ollama.ChatModel
}
