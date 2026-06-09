package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/LouYuanbo1/go-eino-agent/tools/retriever"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type CrawlerAgent struct {
	Agent *adk.ChatModelAgent
}

func NewCrawlerAgent(ctx context.Context, config *adk.ChatModelAgentConfig) *CrawlerAgent {
	agent, err := adk.NewChatModelAgent(ctx, config)
	if err != nil {
		fmt.Printf("Error creating crawler agent: %v", err)
		return nil
	}
	return &CrawlerAgent{Agent: agent}
}

func NewDefaultCrawlerAgent(ctx context.Context, model model.ToolCallingChatModel, crawler Crawler) *CrawlerAgent {

	navigateTool, err := NewNavigateTool(ctx, crawler)
	if err != nil {
		fmt.Printf("Error creating navigate tool: %v", err)
		return nil
	}
	clickTool, err := NewClickTool(ctx, crawler)
	if err != nil {
		fmt.Printf("Error creating click tool: %v", err)
		return nil
	}
	scrollTool, err := NewScrollTool(ctx, crawler)
	if err != nil {
		fmt.Printf("Error creating scroll tool: %v", err)
		return nil
	}
	jsTool, err := NewJavaScriptTool(ctx, crawler)
	if err != nil {
		fmt.Printf("Error creating javascript tool: %v", err)
		return nil
	}
	htmlTool, err := NewHTMLTool(ctx, crawler)
	if err != nil {
		fmt.Printf("Error creating html tool: %v", err)
		return nil
	}

	instruction :=
		`
			## 角色定义
			你是一位专业的网页数据抓取智能体，基于真实浏览器环境自动化操作网页，能够打开任意URL、模拟用户交互、执行JavaScript、提取页面内容。你擅长处理JS动态渲染页面，精准获取用户所需的网页数据。

			## 核心能力
			1. **页面导航**：打开指定URL并等待页面完全加载（包括JS动态内容渲染完成）
			2. **元素交互**：通过CSS选择器或XPath精准点击页面元素，触发内容加载或页面跳转
			3. **页面滚动**：支持指定像素级的页面滚动，用于触发懒加载内容或浏览长页面
			4. **JS脚本执行**：在页面上下文中执行任意JavaScript代码，获取动态数据或操作DOM
			5. **内容提取**：支持两种提取模式——智能正文提取（基于可读性算法自动去除广告/导航等噪音）或完整HTML获取

			## 可用工具详解
			| 工具名 | 用途 | 适用场景 |
			|--------|------|----------|
			| navigate_tool | 导航到目标URL并等待页面加载完成 | 用户提供URL需要查看/抓取页面内容时，作为第一步操作 |
			| click_tool | 通过CSS选择器或XPath点击页面元素 | 需要点击"加载更多"、翻页按钮、展开折叠内容、切换Tab时 |
			| scroll_tool | 按指定像素数滚动页面 | 触发无限滚动加载、浏览长页面、等待懒加载内容出现时 |
			| javascript_tool | 在页面中执行JavaScript并获取返回值 | 提取复杂动态数据、操作DOM、触发自定义交互逻辑时 |
			| html_tool | 获取当前页面的HTML或提取正文 | 已导航到目标页面后，需要提取内容时作为最终步骤使用 |

			## 操作工作流
			典型的数据抓取流程：
			1. **导航页面** → 使用 navigate_tool 导航到目标URL
			2. **交互操作** → 根据需要使用 click_tool / scroll_tool / javascript_tool 触发内容加载
			3. **内容提取** → 使用 html_tool 获取最终内容（默认智能提取正文，必要时获取完整HTML）

			注意：
			- 每次调用工具后会返回页面的最新内容，请据此判断是否需要继续操作
			- html_tool 在不传 EnableFullInfo 时使用可读性算法智能提取正文，适合大多数场景；传 EnableFullInfo=true 时返回完整HTML
			- 对于JS动态渲染页面，请确保 navigate_tool 完成加载后再进行后续操作

			## 行为准则
			- 所有操作基于用户明确指令，不主动访问用户未指定的URL
			- 优先使用智能正文提取模式（EnableFullInfo=false），减少无效信息返回
			- 在点击或滚动后，如页面内容已满足需求，应及时调用 html_tool 提取内容
			- 遇到页面加载失败或元素未找到时，如实报告错误并尝试替代方案
			- 提取的正文内容保持原文结构，不添加个人解读
			- 不访问或提取涉及违法、侵权内容的页面

			## 输出格式规范
			完成抓取任务后，请按以下格式组织结果：

			【页面信息】URL、标题（如可获取）
			【抓取内容】以清晰的段落结构呈现提取的正文内容，保留原标题层级
			【操作摘要】简述执行了哪些工具调用及其结果
			【注意事项】（如适用）页面部分内容未加载、需登录、反爬限制等

			## 错误处理策略
			- 页面加载超时 → 报告超时并建议用户确认URL是否正确
			- 元素选择器未命中 → 尝试提供页面中实际存在的相似元素选择器建议
			- JS执行异常 → 报告具体错误信息，检查JS语法或页面是否已加载目标元素
			- 页面返回空内容 → 说明可能原因（反爬、需登录、纯Canvas渲染等），建议尝试 EnableFullInfo=true

			## 交互示例
			用户：「打开 https://example.com/news 并抓取今天的头条新闻」
			助手操作流程：
			1. 调用 navigate_tool(url="https://example.com/news")
			2. 观察返回内容，识别头条新闻所在区域
			3. 如需滚动查看，调用 scroll_tool(scroll_y=300)
			4. 调用 html_tool(enable_full_info=false) 提取正文
			5. 将提取的新闻内容按【页面信息】【抓取内容】【操作摘要】格式输出给用户
		`
	searchAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "crawlerAgent",
		Description: "一个基于大模型的网页数据抓取智能体，支持浏览器自动化操作与智能内容提取",
		Instruction: instruction,
		Model:       model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					navigateTool,
					clickTool,
					scrollTool,
					jsTool,
					htmlTool,
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("Error creating search agent: %v", err)
		return nil
	}
	return &CrawlerAgent{Agent: searchAgent}
}

func (c *CrawlerAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return c.Agent.Run(ctx, input, options...)
}

func (c *CrawlerAgent) OutputMessage(ctx context.Context, input string, streamFunc func(string), options ...adk.AgentRunOption) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: c.Agent, EnableStreaming: true})
	iter := runner.Query(ctx, input, options...)
	stream(iter, streamFunc)
}

type RetrieverAgent struct {
	Agent *adk.ChatModelAgent
}

func NewRetrieverAgent(ctx context.Context, config *adk.ChatModelAgentConfig) *RetrieverAgent {
	agent, err := adk.NewChatModelAgent(ctx, config)
	if err != nil {
		fmt.Printf("Error creating chat model agent: %v", err)
		return nil
	}
	return &RetrieverAgent{Agent: agent}
}

func NewDefaultRetrieverAgent[R retriever.Retriever](ctx context.Context, model model.ToolCallingChatModel, typedRetriever R) *RetrieverAgent {
	retrieverTool, err := retriever.NewRetrieverTool(ctx, typedRetriever)
	if err != nil {
		fmt.Printf("Error creating retriever tool: %v", err)
		return nil
	}

	instruction :=
		`
		你是一个基于本地知识库的智能问答助手。
		你的主要职责是仅通过查询本地知识库来回答用户的问题，不得依赖外部知识或自行编造信息。请遵循以下指南：
		知识库访问：当用户提出问题时，首先调用本地知识库检索工具（如向量数据库或文件索引），查找与问题最相关的内容片段。如果知识库支持多文档检索，优先覆盖所有相关文档。
		信息整合：根据检索到的知识片段，用清晰、准确的语言组织答案。可以适当引用原文，但不要直接复制大段文本，除非用户要求。如果多个来源信息一致，综合回答；如果有冲突，指出不同观点并说明来源。
		来源标注：在答案末尾附上参考的知识库来源（如果可用），例如作者,URL等。这有助于用户追溯信息。
		无法回答的情况：如果检索后未找到与问题相关的信息，请礼貌地告知用户：“抱歉，我的本地知识库中暂无相关信息。您可以尝试换一种问法，或咨询其他渠道。” 不要试图用通用知识填补空白。
		对话上下文：在多轮对话中，可以结合之前的问题和回答，但每次新的提问仍需基于知识库检索，不可沿用旧回答中未经验证的信息。
		语言风格：使用专业、友好的语气，根据用户的问题复杂度调整解释的详细程度。
		现在，请开始处理用户的问题。
		`
	return NewRetrieverAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "retrieverAgent",
		Description: "一个基于大模型的检索智能体",
		Instruction: instruction,
		Model:       model,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					retrieverTool,
				},
			},
		},
	})
}

func (r *RetrieverAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return r.Agent.Run(ctx, input, options...)
}

func (r *RetrieverAgent) OutputMessage(ctx context.Context, input string, streamFunc func(string), options ...adk.AgentRunOption) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: r.Agent, EnableStreaming: true})
	iter := runner.Query(ctx, input, options...)
	stream(iter, streamFunc)
}

func stream(iter *adk.AsyncIterator[*adk.AgentEvent], streamFunc func(string)) {
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("Runner 错误: %v", event.Err)
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			mo := event.Output.MessageOutput

			// 【关键修复 1】：检查 MessageStream 是否为 nil
			if mo.MessageStream == nil {
				continue // 跳过 nil stream
			}

			// 只有 Stream 不为 nil 时，才进入流式处理
			streamMessage(mo.MessageStream, streamFunc)
		}
	}
}

func streamMessage(s adk.MessageStream, streamFunc func(string)) {
	// 【关键修复 2】：防御性判空，防止外部意外传入 nil 导致 panic
	if s == nil {
		return
	}
	defer s.Close()

	// 工具调用信息缓存：在流式模式下，工具调用参数会跨多个 chunk 到达，需要累积
	accToolCalls := make(map[int]*schema.ToolCall)

	for {
		chunk, err := s.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Stream Recv 错误: %v", err)
			break
		}

		// 处理思考过程 (如果开启了 DeepSeek-R1 等支持 Reasoning 的模型)
		if chunk.ReasoningContent != "" {
			flushToolCalls(accToolCalls, streamFunc)
			streamFunc(chunk.ReasoningContent)
			continue
		}

		// 处理工具调用：累积各 chunk 中的参数片段
		if len(chunk.ToolCalls) > 0 {
			accumulateToolCalls(accToolCalls, chunk.ToolCalls)
			continue
		}

		// 处理文本内容：先刷新累积的工具调用，再输出文本
		if chunk.Content != "" {
			flushToolCalls(accToolCalls, streamFunc)
			streamFunc(chunk.Content)
		}
	}

	// 流结束时刷新残留的工具调用
	flushToolCalls(accToolCalls, streamFunc)
}

// accumulateToolCalls 将流式 chunk 中的工具调用信息累积到 acc 中
func accumulateToolCalls(acc map[int]*schema.ToolCall, toolCalls []schema.ToolCall) {
	for _, tc := range toolCalls {
		idx := 0
		if tc.Index != nil {
			idx = *tc.Index
		}

		if existing, ok := acc[idx]; ok {
			if tc.ID != "" {
				existing.ID = tc.ID
			}
			if tc.Type != "" {
				existing.Type = tc.Type
			}
			if tc.Function.Name != "" {
				existing.Function.Name = tc.Function.Name
			}
			existing.Function.Arguments += tc.Function.Arguments
		} else {
			cp := tc
			acc[idx] = &cp
		}
	}
}

// flushToolCalls 输出所有累积的工具调用信息并清空缓存
func flushToolCalls(acc map[int]*schema.ToolCall, streamFunc func(string)) {
	for idx := 0; idx < len(acc); idx++ {
		tc, ok := acc[idx]
		if !ok {
			continue
		}
		var info strings.Builder
		info.WriteByte('\n')
		fmt.Fprintf(&info, "[工具调用 %d] 名称: %s", idx+1, tc.Function.Name)
		if len(tc.Function.Arguments) > 0 {
			fmt.Fprintf(&info, ", 参数: %s", tc.Function.Arguments)
		}
		info.WriteByte('\n')
		streamFunc(info.String())
		delete(acc, idx)
	}
}

type SupervisorAgent struct {
	Agent adk.Agent
}

func NewSupervisorAgent[R retriever.Retriever](ctx context.Context, model model.ToolCallingChatModel, crawler Crawler, typedRetriever R) (*SupervisorAgent, error) {
	sv, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "the agent responsible to supervise tasks",
		Instruction: `
		You are a supervisor managing two agents:

        - a research agent. Assign research-related tasks to this agent
        - a math agent. Assign math-related tasks to this agent
        Assign work to one agent at a time, do not call agents in parallel.
        Do not do any work yourself.`,
		Model: model,
		Exit:  &adk.ExitTool{},
	})
	if err != nil {
		return nil, err
	}

	crawlerAgent := NewDefaultCrawlerAgent(ctx, model, crawler)
	retrieverAgent := NewDefaultRetrieverAgent(ctx, model, typedRetriever)

	s,err:=supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  []adk.Agent{crawlerAgent.Agent, retrieverAgent.Agent},
	})
	if err!= nil {
		return nil, err
	}
	return &SupervisorAgent{Agent: s}, nil
}

func (s *SupervisorAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return s.Agent.Run(ctx, input, options...)
}

func (s *SupervisorAgent) OutputMessage(ctx context.Context, input string, streamFunc func(string), options ...adk.AgentRunOption) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: s.Agent, EnableStreaming: true})
	iter := runner.Query(ctx, input, options...)
	stream(iter, streamFunc)
}
