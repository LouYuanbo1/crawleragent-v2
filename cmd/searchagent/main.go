package main

import (
	"context"
	"crawleragent-v2/internal/config"
	"crawleragent-v2/internal/domain/model"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/llm"
	"crawleragent-v2/internal/infra/persistence/es"
	"crawleragent-v2/internal/service/searchagent"
	"crawleragent-v2/param"
	_ "embed"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

//使用go:embed嵌入appconfig.json文件
//下方注释重要,不能删除
//在实际使用时，注意与文件名的对应，Github上保存的appconfig_example.json文件为样例，以实际为准,比如我这里是appconfig.json
//When using it in practice, pay attention to the correspondence between the filename and the actual filename.
//The appconfig_example.json file saved on GitHub is just an example;
//use your own file, for example, mine is appconfig.json.

//go:embed appconfig/appconfig.json
var appConfig []byte

func main() {
	appcfg, err := config.ParseConfig(appConfig)
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	fmt.Printf("Rod UserDataDir: %s\n", appcfg.Rod.UserDataDir)

	ctx := context.Background()

	typedClient, err := es.InitTypedEsClient(appcfg, 3)
	if err != nil {
		log.Fatalf("failed to initialize Elasticsearch client: %s", err)
	}

	embedder, err := embedding.InitEmbedder(ctx, appcfg, 5, 1)
	if err != nil {
		log.Fatalf("初始化Embedder失败: %v", err)
	}

	llm, err := llm.InitLLM(ctx, appcfg)
	if err != nil {
		log.Fatalf("初始化LLM失败: %v", err)
	}

	//定义搜索模式的提示模板

	searchModePrompt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`

		角色：你是一位专业的职业顾问，擅长根据用户背景和需求匹配合适的工作岗位，并能灵活利用知识库资源。

		任务：根据用户提供的求职需求，结合ES知识库中的岗位信息，为用户推荐最匹配的工作岗位。当知识库中无相关结果时，转而推荐优质招聘平台。

		输入信息：

		用户需求：包括但不限于行业偏好、工作地点、薪资期望、工作经验、技能要求、学历背景等
		ES知识库内容：从企业知识库中检索到的相关岗位信息（可能为空）
		处理逻辑：

		优先使用知识库：仔细分析Elasticsearch知识库返回的岗位数据，根据用户需求进行精准匹配
		匹配维度：考虑岗位职责、任职要求、薪资范围、工作地点、发展空间等关键因素
		知识库为空时：当知识库未返回任何相关岗位时，立即切换到推荐招聘网站模式
		推荐原则：始终以用户需求为中心，提供实用、可操作的建议
		输出格式：

		当知识库有结果时：
		【精准岗位推荐】
		根据您的需求，为您推荐以下岗位：
		🔹 [岗位名称] - [公司名称]
		• 工作地点：[地点]
		• 薪资范围：[薪资]
		• 核心要求：[2-3个关键要求]
		• 网址：[岗位详情链接]
		• 匹配理由：[说明为何匹配用户需求]
		💡 建议：[1-2条具体求职建议]

		当知识库为空时：
		【招聘平台推荐】
		不要编造任何数据,不要进行任何假设。
		直接回答:当前知识库中暂无完全匹配的岗位，建议您通过以下专业招聘平台自主搜索：
		综合类平台：
		• 智联招聘（www.zhaopin.com）- 覆盖全行业，岗位数量丰富
		• 前程无忧（www.51job.com）- 企业质量高，适合中高端求职
		垂直类平台：
		• 拉勾网（www.lagou.com）- 专注互联网/科技行业
		• 猎聘（www.liepin.com）- 高端职位和猎头服务
		新兴平台：
		• BOSS直聘（www.zhipin.com）- 直接与招聘方沟通，效率高
		• 脉脉（www.maimai.cn）- 职场社交+内推机会
		💡 使用建议：建议在搜索时使用关键词组合（如"行业+职位+地点"），并完善个人简历以提高匹配度。
		注意事项：

		保持推荐的专业性和实用性
		避免推荐明显不匹配的岗位
		当知识库结果较少时，可适当补充招聘网站建议
		用简洁清晰的语言表达，避免过于技术化的术语
		始终以帮助用户成功求职为目标
		`),
		schema.SystemMessage(`以下是根据您的查询检索到的相关岗位信息：\n{referenceDocs}\n\n请严格展示这些岗位编号和信息,不要编造或添加任何额外信息。如果知识库为空,则直接回答:当前知识库中暂无完全匹配的岗位。`),
		schema.UserMessage("{query}"),
	)

	//定义聊天模式的提示模板
	chatModePrompt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`
		角色：你是一个专业的职业顾问，擅长根据用户背景和需求匹配合适的工作岗位，并于用户交流职业规划，倾听用户求助时的烦恼，提供专业的建议并鼓励用户。
		任务：与用户进行互动，倾听用户的问题、需求和建议，根据用户的背景和需求提供专业的职业规划建议。
		`),
		schema.SystemMessage(`以下是根据您经过网络查询得到的信息：\n{duckDuckGoResults}\n\n请结合这些信息回答用户的请求。`),
		schema.UserMessage("{query}"),
	)

	//初始化Agent
	params := &param.Agent{
		Doc: &model.BossJobDoc{},
		Prompt: map[param.PromptType]*prompt.DefaultChatTemplate{
			param.PromptEsRAGMode: searchModePrompt,
			param.PromptChatMode:  chatModePrompt,
		},
	}
	agent, err := searchagent.InitSearchAgent(ctx,
		llm,
		typedClient,
		embedder,
		params)
	if err != nil {
		log.Fatalf("初始化Agent失败: %v", err)
	}
	fmt.Println("欢迎使用CrawlerAgent!")
	fmt.Println("注意:当请求以'查询模式'开头时,会使用Es知识库,否则会默认为聊天模式。")
	fmt.Println("知识库内容越多,描述越完善,推荐结果越准确。")
	fmt.Println("请输入您的请求:")
	query := ""
	for {
		//读取用户输入
		fmt.Scanln(&query)
		fmt.Printf("\n")

		switch query {
		case "":
			continue
		case "exit", "e", "quit", "q":
			fmt.Println("感谢使用CrawlerAgent!")
			return
		}
		//调用Agent,使用流式输出
		err = agent.Stream(ctx, query)
		if err != nil {
			log.Fatalf("调用Agent失败: %v", err)
		}
	}
}
