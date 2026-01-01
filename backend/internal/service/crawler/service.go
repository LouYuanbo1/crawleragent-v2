package service

import (
	"context"
	"crawleragent-v2/internal/data/model"
	"crawleragent-v2/param"
)

type CrawlerService interface {
	StartCrawling(ctx context.Context, params []*param.ParallelCrawlerParam) error
	EmbeddingAndIndexDocs(ctx context.Context, docs []model.Document) error
}
