package parallel

import (
	"context"
)

type ParallelCrawler interface {
	Close()
	Crawl(ctx context.Context, runtimes []*ParallelCrawlerRuntime) error
}
