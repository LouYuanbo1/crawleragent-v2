package main

import (
	"context"
	"crawleragent-v2/config"
	"crawleragent-v2/internal/crawler"
	"crawleragent-v2/internal/embedding"
	"crawleragent-v2/internal/model"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LouYuanbo1/go-webservice/elasticsearchx"
	"github.com/cloudwego/eino-ext/components/embedding/ollama"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
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

	crawl, err := crawler.InitBrowserPoolCrawler(&appcfg.Rod, 3)
	if err != nil {
		log.Fatalf("初始化浏览器池爬虫失败: %v", err)
	}

	ctx := context.Background()

	ebd, err := ollama.NewEmbedder(ctx, &ollama.EmbeddingConfig{
		Model:   appcfg.Embedding.Model,
		BaseURL: fmt.Sprintf("%s:%d", appcfg.Embedding.Host, appcfg.Embedding.Port),
	})

	embedder, err := embedding.InitEmbedder(ctx, ebd, 32, 3)
	if err != nil {
		log.Fatalf("初始化嵌入模型失败: %v", err)
	}
	typedClient, err := elasticsearch.NewTypedClient(elasticsearch.Config{
		Username: appcfg.Elasticsearch.Username,
		Password: appcfg.Elasticsearch.Password,
		Addresses: []string{
			fmt.Sprintf("%s:%d", appcfg.Elasticsearch.Host, appcfg.Elasticsearch.Port),
		},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			// 跳过TLS验证（仅在开发环境中使用）
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize Elasticsearch client: %s", err)
	}

	esx := elasticsearchx.NewElasticsearchX(typedClient)

	clickXAction := &crawler.ClickXAction{
		Delay:    2000 * time.Millisecond,
		Selector: selectorCnBlogs,
	}

	scrollAction := &crawler.ScrollAction{
		Delay:   2000 * time.Millisecond,
		ScrollY: 1000,
	}

	jsAction := &crawler.JavaScriptAction{

		Delay: 2000 * time.Millisecond,
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
		ProcessFunc: func(ctx context.Context, content crawler.UrlContent) error {
			log.Printf("执行JavaScript成功:%s, %d, %s", content.GetUrl(), len(content.GetContent()), string(content.GetContent()[:100]))
			return nil
		},
	}

	scrollAndJsActions := make([]crawler.Action, 0, 10)
	clickXAndJsActions := make([]crawler.Action, 0, 10)

	for range 5 {
		scrollAndJsActions = append(scrollAndJsActions, scrollAction)
		scrollAndJsActions = append(scrollAndJsActions, jsAction)
	}

	for range 5 {
		clickXAndJsActions = append(clickXAndJsActions, clickXAction)
		clickXAndJsActions = append(clickXAndJsActions, jsAction)
	}

	processFuncBoss := func(ctx context.Context, content crawler.UrlContent) error {
		var jsonData struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			ZpData  struct {
				HasMore    bool                   `json:"hasMore"`
				JobResList []model.RowBossJobData `json:"jobList"`
			} `json:"zpData"`
		}

		if err := json.Unmarshal(content.GetContent(), &jsonData); err != nil {
			return fmt.Errorf("JSON解析失败: %v", err)
		}

		if jsonData.Code != 0 {
			return fmt.Errorf("API返回错误: %d - %s", jsonData.Code, jsonData.Message)
		}

		results := make([]*model.BossJobDoc, 0, len(jsonData.ZpData.JobResList))
		for _, job := range jsonData.ZpData.JobResList {
			rowData := &model.RowBossJobData{
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

		embeddingStrings := make([]string, 0, len(results))
		for _, doc := range results {
			embeddingStrings = append(embeddingStrings, doc.GetEmbeddingString())
		}
		embeddings, err := embedder.Embed(ctx, embeddingStrings)
		if err != nil {
			return fmt.Errorf("嵌入文档失败: %w", err)
		}
		for i, doc := range results {
			doc.SetEmbedding(embeddings[i])
		}
		err = esx.BulkIndexDocs[model.BossJobDoc](ctx, results, esutil.BulkIndexerConfig{}, true)
		if err != nil {
			return fmt.Errorf("索引文档失败: %w", err)
		}
		log.Printf("转换文档: %v", results)

		return nil
	}

	crawl.Crawl(context.Background(), []*crawler.CrawlerParam{
		{
			URL: urlBoss,
			NetworkConfigs: []*crawler.NetworkConfig{
				{
					URLPattern:  urlPatternBoss,
					ProcessFunc: processFuncBoss,
				},
			},
			Actions: scrollAndJsActions,
		},
		{
			URL: urlCnBlogs,
			NetworkConfigs: []*crawler.NetworkConfig{
				{
					URLPattern: urlPatternCnBlogs,
					ProcessFunc: func(ctx context.Context, content crawler.UrlContent) error {
						log.Printf("执行JavaScript成功:%s, %d, %s", content.GetUrl(), len(content.GetContent()), string(content.GetContent()[:100]))
						return nil
					},
				},
			},
			Actions: clickXAndJsActions,
		},
		{
			URL: urlBili,
			NetworkConfigs: []*crawler.NetworkConfig{
				{
					URLPattern: urlPatternBili,
					ProcessFunc: func(ctx context.Context, content crawler.UrlContent) error {
						log.Printf("执行JavaScript成功:%s, %d, %s", content.GetUrl(), len(content.GetContent()), string(content.GetContent()[:100]))
						return nil
					},
				},
			},
			Actions: scrollAndJsActions,
		},
	})

}
