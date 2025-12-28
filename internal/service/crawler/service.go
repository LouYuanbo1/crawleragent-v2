package crawler

import (
	"context"
	"crawleragent-v2/param"
)

type CrawlerService interface {
	StartCrawling(ctx context.Context, params []*param.ParallelCrawlerParam) error
}
