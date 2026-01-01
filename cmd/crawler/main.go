package main

import (
	"context"
	"crawleragent-v2/internal/config"
	"crawleragent-v2/internal/data/entity"
	"crawleragent-v2/internal/data/model"
	"crawleragent-v2/internal/infra/crawler/parallel"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/persistence/es"
	"crawleragent-v2/internal/service/crawler"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var (
	urlBoss           = "https://www.zhipin.com/web/geek/jobs?city=100010000&salary=406&experience=102&query=golang"
	urlPatternBoss    = "https://www.zhipin.com/wapi/zpgeek/search/joblist.json*"
	urlCnBlogs        = "https://www.cnblogs.com/"
	urlPatternCnBlogs = "https://www.cnblogs.com/AggSite/AggSitePostList*"
	selectorCnBlogs   = `//a[starts-with(@href, "/sitehome/p/") and text()=">"]`
	urlBili           = "https://www.bilibili.com/"
	urlPatternBili    = "https://api.bilibili.com/x/web-interface/index/ogv/rcmd*"
	urlCsdn           = "https://www.csdn.net/"
	urlPatternCsdn    = "https://cms-api.csdn.net/v1/web_home/select_content*"
)

func main() {
	appcfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	fmt.Printf("Chromedp UserDataDir: %s\n", appcfg.Rod.UserDataDir)

	//context.Background()
	// 这是最常用的根Context，通常用在main函数、初始化或测试中，作为整个Context树的顶层。
	// 当你不知道使用哪个Context，或者没有可用的Context时，可以使用它作为起点。
	// 它永远不会被取消，没有超时时间，也没有值。
	ctx := context.Background()
	//运行前确保es服务启动完成
	parallelCrawler, err := parallel.InitBrowserPoolCrawler(appcfg, 3)
	if err != nil {
		log.Fatalf("初始化BrowserPoolCrawler失败: %v", err)
	}
	defer parallelCrawler.Close()

	clickXActions := make([]param.Action, 0, 5)
	for range 5 {
		clickXActions = append(clickXActions, &param.ClickXAction{
			BaseParams: param.BaseParams{
				Delay: 2000 * time.Millisecond,
			},
			Selector: selectorCnBlogs,
		},
		)
	}

	scrollActions := make([]param.Action, 0, 5)
	for range 5 {
		scrollActions = append(scrollActions, &param.ScrollAction{
			BaseParams: param.BaseParams{
				Delay: 2000 * time.Millisecond,
			},
			ScrollY: 1000,
		})
	}

	jsActions := make([]param.Action, 0, 5)
	for range 5 {
		jsActions = append(jsActions, &param.JavaScriptAction{
			BaseParams: param.BaseParams{
				Delay: 2000 * time.Millisecond,
			},
			JavaScript: `
					() => {
						function getAllHrefLinks() {
							// 选择器「[href]」匹配所有拥有href属性的元素（无论标签类型）
							const hrefElements = document.querySelectorAll('[href]');
							
							const allHrefs = Array.from(hrefElements)
								.map(element => {
								// 按需选择：获取绝对路径 OR 原始属性值
								const absoluteHref = element.href; // 完整URL（推荐，通用性更强）
								const originalHref = element.getAttribute('href'); // 原始值（如 "../test.html"、"#top"）
								return absoluteHref; // 可替换为 originalHref
								})
								.filter(href => !!href.trim()) // 过滤无效空链接
								.filter((href, index, self) => self.indexOf(href) === index); // 数组去重
							
							return allHrefs;
						}
						return getAllHrefLinks();
					}
				`,
			ProcessFunc: func(ctx context.Context, content types.UrlContent) error {
				log.Printf("执行JavaScript成功:%s, %d, %s", content.GetUrl(), len(content.GetContent()), string(content.GetContent()[:100]))
				return nil
			},
		})
	}

	scrollAndJsActions := make([]param.Action, 0, 10)
	clickXAndJsActions := make([]param.Action, 0, 10)

	for i := range 5 {
		scrollAndJsActions = append(scrollAndJsActions, scrollActions[i])
		scrollAndJsActions = append(scrollAndJsActions, jsActions[i])
	}

	for i := range 5 {
		clickXAndJsActions = append(clickXAndJsActions, clickXActions[i])
		clickXAndJsActions = append(clickXAndJsActions, jsActions[i])
	}

	typedClient, err := es.InitTypedEsClient(appcfg, 10)
	if err != nil {
		log.Fatalf("初始化TypedEsClient失败: %v", err)
	}

	embedder, err := embedding.InitEmbedder(ctx, appcfg, 5, 1)
	if err != nil {
		log.Fatalf("初始化嵌入器失败: %v", err)
	}

	crawlerService := crawler.InitCrawlerService(parallelCrawler, embedder, typedClient, 5)

	processFuncBoss := func(ctx context.Context, content types.UrlContent) error {
		var jsonData struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			ZpData  struct {
				HasMore    bool                    `json:"hasMore"`
				JobResList []entity.RowBossJobData `json:"jobList"`
			} `json:"zpData"`
		}

		if err := json.Unmarshal(content.GetContent(), &jsonData); err != nil {
			return fmt.Errorf("JSON解析失败: %v", err)
		}

		if jsonData.Code != 0 {
			return fmt.Errorf("API返回错误: %d - %s", jsonData.Code, jsonData.Message)
		}

		results := make([]model.Document, 0, len(jsonData.ZpData.JobResList))
		for _, job := range jsonData.ZpData.JobResList {
			rowData := &entity.RowBossJobData{
				EncryptJobId:     job.EncryptJobId,
				SecurityId:       job.SecurityId,
				JobName:          job.JobName,
				SalaryDesc:       job.SalaryDesc,
				BrandName:        job.BrandName,
				BrandScaleName:   job.BrandScaleName,
				CityName:         job.CityName,
				AreaDistrict:     job.AreaDistrict,
				BusinessDistrict: job.BusinessDistrict,
				JobLabels:        job.JobLabels,
				Skills:           job.Skills,
				JobExperience:    job.JobExperience,
				JobDegree:        job.JobDegree,
				WelfareList:      job.WelfareList,
			}
			doc := rowData.ToDocument()
			results = append(results, doc)
		}

		crawlerService.EmbeddingAndIndexDocs(ctx, results)

		return nil
	}

	params := []*param.ParallelCrawlerParam{
		{
			URL: urlBoss,
			NetworkConfigs: []*param.ParallelNetworkConfig{
				{
					URLPattern:  urlPatternBoss,
					ProcessFunc: processFuncBoss,
				},
			},
			Actions: scrollAndJsActions,
		},

		{
			URL: urlBili,
			NetworkConfigs: []*param.ParallelNetworkConfig{
				{
					URLPattern: urlPatternBili,
				},
			},
			Actions: scrollAndJsActions,
		},
		{
			URL: urlCnBlogs,
			NetworkConfigs: []*param.ParallelNetworkConfig{
				{
					URLPattern: urlPatternCnBlogs,
				},
			},
			Actions: clickXAndJsActions,
		},
		{
			URL: urlCsdn,
			NetworkConfigs: []*param.ParallelNetworkConfig{
				{
					URLPattern: urlPatternCsdn,
				},
			},
			Actions: scrollAndJsActions,
		},
	}
	err = crawlerService.StartCrawling(ctx, params)
	if err != nil {
		log.Fatalf("启动爬虫失败: %v", err)
	}

	count, err := typedClient.CountDocs(ctx, (&model.BossJobDoc{}).GetIndex())
	if err != nil {
		log.Fatalf("查询索引文档数量失败: %v", err)
	}
	//打印索引中的文档数量
	fmt.Printf("索引中的文档数量: %d\n", count)

	err = typedClient.ToExcel(ctx, "C:/Users/15325/Desktop/boss_jobs.xlsx", (&model.BossJobDoc{}).GetIndex(), []string{"salaryDesc"}, 1000)
	if err != nil {
		log.Fatalf("导出索引文档到Excel失败: %v", err)
	}

	log.Println("所有任务完成")
}
