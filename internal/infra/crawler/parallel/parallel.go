package parallel

import (
	"context"
	"crawleragent-v2/param"
)

type ParallelCrawler interface {
	Close()
	Crawl(ctx context.Context, params []*param.ParallelCrawlerParam) error
}
