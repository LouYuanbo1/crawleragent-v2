package service

import (
	"context"
	"crawleragent-v2/internal/data/model"
	"crawleragent-v2/internal/infra/crawler/parallel"
	"crawleragent-v2/internal/infra/embedding"
	"crawleragent-v2/internal/infra/persistence/es"
	"crawleragent-v2/param"
	"fmt"
	"log"
	"time"

	"github.com/panjf2000/ants/v2"
)

type crawlerService struct {
	parallelCrawler parallel.ParallelCrawler
	taskPool        *ants.Pool
	embedder        embedding.Embedder
	typedClient     es.TypedEsClient
}

func InitCrawlerService(parallelCrawler parallel.ParallelCrawler, embedder embedding.Embedder, typedClient es.TypedEsClient, sizePool int) CrawlerService {
	taskPool, err := ants.NewPool(sizePool)
	if err != nil {
		log.Fatalf("初始化任务池失败: %v", err)
	}
	return &crawlerService{
		parallelCrawler: parallelCrawler,
		taskPool:        taskPool,
		embedder:        embedder,
		typedClient:     typedClient,
	}
}

func (c *crawlerService) StartCrawling(ctx context.Context, params []*param.ParallelCrawlerParam) error {
	err := c.parallelCrawler.Crawl(ctx, params)
	if err != nil {
		return fmt.Errorf("并行爬虫运行失败: %v", err)
	}
	return nil
}

func (c *crawlerService) EmbeddingAndIndexDocs(ctx context.Context, docs []model.Document) error {
	// 为嵌入和索引文档添加超时, 20秒。将来会改成从配置文件读取。
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

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
