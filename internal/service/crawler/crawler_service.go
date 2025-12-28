package crawler

import (
	"context"
	"crawleragent-v2/internal/domain/model"
	"crawleragent-v2/internal/infra/crawler/parallel"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/persistence/es"

	"crawleragent-v2/param"
	"crawleragent-v2/types"
	"log"
	"sync"
)

type crawlerService struct {
	parallelCrawler  parallel.ParallelCrawler
	networkRespChans []chan *types.NetworkResponse
	chanWg           sync.WaitGroup
	embedder         embedding.Embedder
	typedClient      es.TypedEsClient
}

func InitCrawlerService(parallelCrawler parallel.ParallelCrawler, embedder embedding.Embedder, typedClient es.TypedEsClient) CrawlerService {
	return &crawlerService{
		parallelCrawler:  parallelCrawler,
		networkRespChans: make([]chan *types.NetworkResponse, 0),
		embedder:         embedder,
		typedClient:      typedClient,
	}
}

func (c *crawlerService) StartCrawling(ctx context.Context, params []*param.ParallelCrawlerParam) error {
	var runtimes []*parallel.ParallelCrawlerRuntime
	for _, param := range params {
		runtime := c.initChan(ctx, param)
		runtimes = append(runtimes, runtime)
	}
	err := c.parallelCrawler.Crawl(ctx, runtimes)
	if err != nil {
		return err
	}
	for _, respChan := range c.networkRespChans {
		close(respChan)
	}
	c.chanWg.Wait()
	return nil
}

func (c *crawlerService) initChan(ctx context.Context, param *param.ParallelCrawlerParam) *parallel.ParallelCrawlerRuntime {
	runtime := &parallel.ParallelCrawlerRuntime{
		URL:     param.URL,
		Actions: param.Actions,
	}
	var networkConfigs []*parallel.ParallelNetworkRuntime
	for _, networkConfig := range param.NetworkConfigs {
		respChan := make(chan *types.NetworkResponse, networkConfig.RespChanSize)
		networkConfigs = append(networkConfigs, &parallel.ParallelNetworkRuntime{
			URLPattern: networkConfig.URLPattern,
			RespChan:   respChan,
		})
		c.networkRespChans = append(c.networkRespChans, respChan)
		c.chanWg.Add(1)
		go c.processRespChan(ctx, respChan, networkConfig.ToDocFunc)
	}
	runtime.NetworkConfigs = networkConfigs

	return runtime
}

func (c *crawlerService) processRespChan(ctx context.Context, respChan <-chan *types.NetworkResponse, toDocFunc func(body []byte) ([]model.Document, error)) {
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
				docs, err := toDocFunc([]byte(resp.Body))
				if err != nil {
					continue
				}
				log.Printf("处理响应: %s", resp.Url)
				embeddingStrings := make([]string, 0, len(docs))
				for _, doc := range docs {
					embeddingStrings = append(embeddingStrings, doc.GetEmbeddingString())
				}
				//log.Printf("词嵌入语句: %v", embeddingStrings)
				embeddings, err := c.embedder.Embed(ctx, embeddingStrings)
				if err != nil {
					continue
				}
				for i, doc := range docs {
					doc.SetEmbedding(embeddings[i])
				}
				err = c.typedClient.BulkIndexDocsWithID(ctx, docs)
				if err != nil {
					continue
				}
				log.Printf("转换文档: %v", docs)
			} else {
				log.Printf("处理响应: %s, URLPattern: %s, 长度: %d", resp.Url, resp.UrlPattern, len(resp.Body))
			}
		}
	}
}
