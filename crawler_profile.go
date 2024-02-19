package collyresponsible

import (
	"context"
	"time"

	"github.com/gocolly/colly/v2"
)

type CrawlerProfile struct {
	Ctx       context.Context
	Website   string
	UserAgent string
	// Limits
	MaxDepth   int
	MaxRuntime time.Duration
	// Colly configuration
	CollyOptions []colly.CollectorOption
	CollyLimits  *colly.LimitRule
	// Custom callbacks
	ResponseHooks []func(response *colly.Response)
	URLTests      []func(url string) bool
	URLHooks      []func(url string)
}
