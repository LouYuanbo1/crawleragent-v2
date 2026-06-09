# CrawlerAgent-v2

CrawlerAgent-v2 是一个基于 AI Agent 架构的智能网页爬虫系统，采用 Supervisor 模式编排多个专业 Agent，通过大语言模型驱动浏览器自动化操作、智能内容提取和本地知识库检索。

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                    Supervisor Agent                      │
│               (任务路由与子Agent协调)                       │
├─────────────────────┬───────────────────────────────────┤
│   Crawler Agent     │       Retriever Agent              │
│  (浏览器自动化抓取)    │     (本地知识库检索)                 │
├─────────────────────┼───────────────────────────────────┤
│  navigate_tool      │       retriever_tool               │
│  click_tool         │       (Elasticsearch 向量搜索)      │
│  scroll_tool        │                                    │
│  javascript_tool    │                                    │
│  html_tool          │                                    │
└─────────────────────┴───────────────────────────────────┘
```

## 项目特点

- **Supervisor 多 Agent 编排**：基于 CloudWeGo Eino ADK 和 prebuilt/supervisor，由一个 Supervisor Agent 自动判断任务类型并分发给 Crawler Agent 或 Retriever Agent 执行
- **LLM 驱动的浏览器自动化**：Agent 自主决定使用何种工具（导航、点击、滚动、执行 JS、提取内容），无需手动编排操作序列
- **智能正文提取**：基于 go-readability 算法，自动去除广告、导航栏等噪音，提取页面核心内容
- **浏览器池并发爬取**：支持多浏览器实例并发执行，每个实例独立运行，配合 Worker 池高效处理批量任务
- **网络响应拦截**：实时捕获匹配特定 URL 模式的 API 响应，适合抓取 XHR/Fetch 接口数据
- **向量检索**：集成 Elasticsearch + Ollama Embedding，支持本地知识库的语义搜索
- **REST API 服务**：基于 Gin 提供 HTTP 接口，可集成到其他系统中
- **Vue 3 前端**：Vite + TypeScript + Vue 3 构建的现代化前端界面

## 技术栈

| 层级 | 技术 |
|------|------|
| AI Agent 框架 | CloudWeGo Eino (ADK + Supervisor) |
| LLM | DeepSeek API / Ollama 本地模型 |
| 浏览器自动化 | go-rod + go-rod/stealth |
| 内容提取 | go-readability |
| 向量嵌入 | Ollama (nomic-embed-text) |
| 搜索引擎 | Elasticsearch 9.x |
| HTTP 框架 | Gin |
| 配置管理 | Viper |
| 前端 | Vue 3 + TypeScript + Vite + Axios |
| 测试 | testify |

## 目录结构

```
crawleragent-v2/
├── backend/
│   ├── cmd/
│   │   ├── agent/             # Agent 主程序入口（Supervisor模式）
│   │   ├── crawler/           # 批量爬虫入口
│   │   └── gui/               # Web GUI 服务入口
│   ├── config/
│   │   ├── config.go          # 配置结构定义与加载
│   │   └── config_example.yaml# 配置示例文件
│   ├── internal/
│   │   ├── agent/
│   │   │   ├── agent.go       # Agent 定义：CrawlerAgent、RetrieverAgent、SupervisorAgent
│   │   │   ├── crawler.go     # 浏览器爬虫实现（rod）
│   │   │   ├── crawltool.go   # Agent 工具定义（navigate/click/scroll/js/html）
│   │   │   ├── option.go      # 浏览器启动器选项
│   │   │   └── option_test.go # 启动器选项测试
│   │   ├── crawler/
│   │   │   ├── crawler.go     # 浏览器池爬虫实现
│   │   │   ├── action.go      # 页面操作定义（Click/Scroll/JS）
│   │   │   ├── param.go       # 爬虫参数与网络配置
│   │   │   ├── crawler_test.go
│   │   │   ├── action_test.go
│   │   │   ├── option_test.go
│   │   │   └── param_test.go
│   │   ├── controller/
│   │   │   ├── agent/
│   │   │   │   ├── agent.go       # Agent HTTP 控制器
│   │   │   │   ├── dto.go         # 请求/响应 DTO
│   │   │   │   └── middleware.go  # 配置注入中间件
│   │   │   └── document/
│   │   │       └── document.go    # ES 文档查询控制器
│   │   ├── embedding/
│   │   │   └── embedding.go   # 向量嵌入器（批量+信号量控制）
│   │   └── model/
│   │       └── model.go       # 数据模型（BossJobDoc 等）
│   └── go.mod
├── frontend/
│   ├── src/                   # Vue 3 源码（开发中）
│   ├── package.json           # 前端依赖
│   ├── vite.config.ts         # Vite 配置（含 API 代理）
│   └── index.html
├── .gitignore
├── LICENSE
└── README.md
```

## 核心组件

### 1. SupervisorAgent

系统的总调度器，基于 CloudWeGo Eino 的 `prebuilt/supervisor`，管理两个子 Agent：

- **CrawlerAgent**：负责网页抓取相关任务（导航、交互、内容提取）
- **RetrieverAgent**：负责本地知识库检索任务（向量语义搜索）

Supervisor 接收用户自然语言指令后，自动判断应交给哪个子 Agent 处理。

### 2. CrawlerAgent

网页抓取智能体，配备 5 个工具：

| 工具 | 功能 | 典型用法 |
|------|------|----------|
| `navigate_tool` | 导航到目标 URL | 打开任意网页 |
| `click_tool` | CSS/XPath 元素点击 | 翻页、展开、切换 Tab |
| `scroll_tool` | 页面滚动 | 触发懒加载、无限滚动 |
| `javascript_tool` | 执行 JavaScript | 提取动态数据、操作 DOM |
| `html_tool` | 获取页面内容 | 智能正文提取或完整 HTML |

每个工具执行后自动返回页面最新内容，Agent 可根据返回结果判断是否需要继续操作。

### 3. RetrieverAgent

本地知识库检索智能体，通过 Elasticsearch 向量搜索从已索引的文档中检索相关信息。适用于：
- 搜索已爬取的结构化数据（如招聘信息）
- 基于语义相似度的模糊查询

### 4. 浏览器池爬虫（crawler）

独立的批量爬虫实现，支持：

- **多实例并发**：预启动多个浏览器实例，通过 Worker 池分发任务
- **网络拦截**：通过 URL 模式匹配捕获 XHR/Fetch 响应
- **操作链**：支持 Click、Scroll、JavaScript 等多种操作的顺序组合
- **回调处理**：每个操作和网络拦截都可注册自定义处理函数

## 快速开始

### 前置要求

- Go 1.27+
- Node.js 18+（前端）
- Chrome / Chromium 浏览器
- Ollama（用于本地嵌入模型）
- Elasticsearch 9.x
- DeepSeek API Key（或 Ollama 本地 LLM）

### 1. 安装与配置

```bash
# 克隆项目
git clone https://github.com/yourusername/crawleragent-v2.git
cd crawleragent-v2/backend
```

复制配置示例并修改：

```bash
cp config/config_example.yaml config/config.yaml
```

编辑 `config/config.yaml`，填入你的配置：

```yaml
elasticsearch:
  username: your_es_username
  password: your_es_password
  host: http://localhost
  port: 9200

rod:
  user_data_dir: user_data_dir
  headless: false           # 调试时可设为 false 以观察浏览器操作
  bin: C:/Program Files/Google/Chrome/Application/chrome.exe

embedding:
  host: http://localhost
  port: 11434
  model: nomic-embed-text

deepseek:
  api_key: your_deepseek_api_key

llm:
  host: http://localhost
  port: 11434
  model: qwen3:1.7b
```

### 2. 启动依赖服务

```bash
# 启动 Elasticsearch
# 启动 Ollama 并拉取模型
ollama pull nomic-embed-text
ollama pull qwen3:1.7b
```

### 3. 运行

```bash
# 方式一：命令行 Agent（Supervisor 模式）
go run cmd/agent/main.go

# 方式二：批量爬虫
go run cmd/crawler/main.go

# 方式三：Web GUI 服务（启动后访问 http://localhost:8080）
go run cmd/gui/main.go
```

### 4. 启动前端（可选）

```bash
cd frontend
npm install
npm run dev
```

前端通过 Vite 代理将 `/api` 请求转发到后端 `http://localhost:8080`。

## 使用示例

### Agent 模式（自然语言驱动）

```go
// 示例：让 Agent 自动抓取 CSDN 首页文章
agent.OutputMessage(ctx, "打开CSDN，向下滑动五次，之后获取主页前三个文章标题", func(s string) {
    fmt.Print(s)
})

// 示例：搜索本地知识库
agent.OutputMessage(ctx, "帮我使用本地搜索寻找一下最近的岗位信息", func(s string) {
    fmt.Print(s)
})
```

Agent 会自动分析任务，选择合适的工具序列执行。

### 批量爬虫模式（编程式控制）

```go
// 初始化浏览器池（3个实例）
crawl, _ := crawler.InitBrowserPoolCrawler(&appcfg.Rod, 3)

// 定义滚动操作
scrollAction := &crawler.ScrollAction{
    Delay:   2000 * time.Millisecond,
    ScrollY: 1000,
}

// 定义 JS 执行操作
jsAction := &crawler.JavaScriptAction{
    Delay: 2000 * time.Millisecond,
    JavaScript: `() => { return document.querySelectorAll('[href]').length; }`,
    ProcessFunc: func(ctx context.Context, content crawler.UrlContent) error {
        log.Printf("链接数量: %s", string(content.GetContent()))
        return nil
    },
}

// 执行爬取任务
crawl.Crawl(context.Background(), []*crawler.CrawlerParam{
    {
        URL: "https://www.zhipin.com/web/geek/jobs?query=golang",
        NetworkConfigs: []*crawler.NetworkConfig{
            {
                URLPattern:  "https://www.zhipin.com/wapi/zpgeek/search/joblist.json*",
                ProcessFunc: processFuncBoss, // 自定义处理 + 写入ES
            },
        },
        Actions: []crawler.Action{scrollAction, scrollAction, jsAction},
    },
})
```

### REST API 调用

```bash
# Agent 测试接口
curl -X POST http://localhost:8080/api/searchagent/test \
  -H "Content-Type: application/json" \
  -d '{"query": "打开 https://www.csdn.net/ 获取首页文章标题"}'

# 文档查询接口
curl "http://localhost:8080/api/documents/boss_jobs?page=1&size=10"

# 索引统计
curl "http://localhost:8080/api/documents/indices"
```

## 配置说明

### 浏览器配置（rod）

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `user_data_dir` | Chrome 用户数据目录 | `user_data_dir` |
| `headless` | 无头模式 | `false` |
| `disable_blink_features` | 禁用 Blink 特性（如 AutomationControlled） | - |
| `incognito` | 隐身模式 | `false` |
| `leakless` | 防泄漏模式 | `true` |
| `bin` | Chrome 可执行文件路径 | - |
| `user_agent` | 自定义 User-Agent | - |

### LLM 配置

| 参数 | 说明 |
|------|------|
| `deepseek.api_key` | DeepSeek API Key |
| `llm.host` | Ollama 地址 |
| `llm.port` | Ollama 端口 |
| `llm.model` | Ollama 模型名 |

### 嵌入模型配置

| 参数 | 说明 |
|------|------|
| `embedding.host` | Ollama 地址 |
| `embedding.port` | Ollama 端口 |
| `embedding.model` | 嵌入模型名（推荐 nomic-embed-text） |

## 应用场景

- **招聘信息采集**：批量抓取招聘网站数据，存入 ES 后进行语义搜索
- **内容监控**：定期抓取目标网页，追踪内容变化
- **竞品分析**：自动采集竞品网站的产品信息、价格、评价
- **知识库构建**：将爬取的网页内容索引到 Elasticsearch，构建可搜索的知识库
- **API 数据捕获**：拦截前端 API 调用，获取结构化 JSON 数据

## 运行测试

```bash
cd backend
go test ./...
```

## 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件

## 联系方式

- 项目地址：https://github.com/yourusername/crawleragent-v2
- 问题反馈：https://github.com/yourusername/crawleragent-v2/issues