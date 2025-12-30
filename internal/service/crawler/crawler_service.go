package crawler

import (
	"context"
	"crawleragent-v2/internal/domain/model"
	"crawleragent-v2/internal/infra/crawler/parallel"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/persistence/es"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"fmt"
	"log"
	"sync"
	"time"
)

type crawlerService struct {
	parallelCrawler parallel.ParallelCrawler
	urlContentChans []chan types.UrlContent
	chanWg          sync.WaitGroup
	embedder        embedding.Embedder
	typedClient     es.TypedEsClient
}

func InitCrawlerService(parallelCrawler parallel.ParallelCrawler, embedder embedding.Embedder, typedClient es.TypedEsClient) CrawlerService {
	return &crawlerService{
		parallelCrawler: parallelCrawler,
		urlContentChans: make([]chan types.UrlContent, 0),
		embedder:        embedder,
		typedClient:     typedClient,
	}
}

func (c *crawlerService) StartCrawling(ctx context.Context, params []*param.ParallelCrawlerParam) error {
	var runtimes []*parallel.ParallelCrawlerRuntime
	for _, param := range params {
		runtime, err := c.initChan(ctx, param)
		if err != nil {
			return fmt.Errorf("初始化通道失败: %v", err)
		}
		runtimes = append(runtimes, runtime)
	}
	err := c.parallelCrawler.Crawl(ctx, runtimes)
	if err != nil {
		return fmt.Errorf("并行爬虫运行失败: %v", err)
	}
	for _, respChan := range c.urlContentChans {
		close(respChan)
	}
	c.chanWg.Wait()
	return nil
}

func (c *crawlerService) initChan(ctx context.Context, params *param.ParallelCrawlerParam) (*parallel.ParallelCrawlerRuntime, error) {
	runtime := &parallel.ParallelCrawlerRuntime{
		URL: params.URL,
	}

	var networkConfigs []*parallel.ParallelNetworkRuntime
	for _, networkConfig := range params.NetworkConfigs {
		respChan := make(chan types.UrlContent, networkConfig.RespChanSize)
		networkConfigs = append(networkConfigs, &parallel.ParallelNetworkRuntime{
			URLPattern: networkConfig.URLPattern,
			RespChan:   respChan,
		})
		c.urlContentChans = append(c.urlContentChans, respChan)
		c.chanWg.Add(1)
		go c.processChan(ctx, respChan, networkConfig.ToDocFunc)
	}
	runtime.NetworkConfigs = networkConfigs

	var runtimeActions []param.Action
	for i := range params.Actions {
		err := params.Actions[i].Validate()
		if err != nil {
			return nil, fmt.Errorf("操作参数校验失败: %v", err)
		}
		if jsAction, ok := params.Actions[i].(*param.JavaScriptAction); ok {
			contentChan := make(chan types.UrlContent, jsAction.ContentChanSize)
			action := &parallel.JavaScriptActionRuntime{
				BaseParams:     jsAction.BaseParams,
				JavaScript:     jsAction.JavaScript,
				JavaScriptArgs: jsAction.JavaScriptArgs,
				ContentChan:    contentChan,
			}
			c.urlContentChans = append(c.urlContentChans, contentChan)
			c.chanWg.Add(1)
			go c.processChan(ctx, contentChan, jsAction.ToDocFunc)
			runtimeActions = append(runtimeActions, action)
		} else {
			runtimeActions = append(runtimeActions, params.Actions[i])
		}
	}

	runtime.Actions = runtimeActions

	return runtime, nil
}

func (c *crawlerService) processChan(ctx context.Context, respChan <-chan types.UrlContent, toDocFunc func(body []byte) ([]model.Document, error)) {
	defer c.chanWg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-respChan:
			if !ok {
				return
			}
			if toDocFunc != nil {
				err := c.embeddingAndIndexDocs(ctx, resp, toDocFunc)
				if err != nil {
					log.Printf("处理响应: %s, URLPattern: %s, 错误: %v", resp.GetUrl(), resp.GetUrlPattern(), err)
					continue
				}
			} else {
				log.Printf("响应: %s, URLPattern: %s, 长度: %d", resp.GetUrl(), resp.GetUrlPattern(), len(resp.GetBody()))
			}
		}
	}
}

func (c *crawlerService) embeddingAndIndexDocs(ctx context.Context, resp types.UrlContent, toDocFunc func(body []byte) ([]model.Document, error)) error {
	// 为嵌入和索引文档添加超时, 20秒。将来会改成从配置文件读取。
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	docs, err := toDocFunc(resp.GetBody())
	if err != nil {
		return err
	}
	log.Printf("处理响应: %s", resp.GetUrl())
	embeddingStrings := make([]string, 0, len(docs))
	for _, doc := range docs {
		embeddingStrings = append(embeddingStrings, doc.GetEmbeddingString())
	}
	embeddings, err := c.embedder.Embed(ctx, embeddingStrings)
	if err != nil {
		return fmt.Errorf("嵌入文档失败: %w", err)
	}
	for i, doc := range docs {
		doc.SetEmbedding(embeddings[i])
	}
	err = c.typedClient.BulkIndexDocsWithID(ctx, docs)
	if err != nil {
		return fmt.Errorf("索引文档失败: %w", err)
	}
	log.Printf("转换文档: %v", docs)
	return nil
}
