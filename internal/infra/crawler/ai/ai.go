package ai

import (
	"context"
	"crawleragent-v2/param"
	"crawleragent-v2/types"
)

type AICrawler interface {
	CloseAll() error
	CloseRouter() error
	NavigateURL(url string) error
	ExecuteActions(actions []param.Action, waitIncludes, waitExcludes []string) error
	GetHTML() (string, error)
	CleanHTML(html string, candidates, includeTags, excludeTags []string) (string, error)
	SetListener(ctx context.Context, urlPatterns []string, respCh chan *types.NetworkResponse)
	RouterRun()
}
