package collyresponsible

import (
	"context"

	"github.com/gocolly/colly/v2"
)

type CrawlerProfile struct {
	Ctx       context.Context
	Website   string
	UserAgent string
	// Limits
	MaxDepth   int
	MaxRuntime int
	// Colly configuration
	CollyOptions []colly.CollectorOption
	CollyLimits  *colly.LimitRule
	// Custom callbacks
	ResponseHooks []func(response *colly.Response)
}
