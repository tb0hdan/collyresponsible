package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	collyresponsible "github.com/tb0hdan/colly-responsible"
)

func main() {
	httpCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	profile := &collyresponsible.CrawlerProfile{
		Ctx:       httpCtx,
		Website:   "https://en.wikipedia.org",
		UserAgent: "Mozilla/5.0 (compatible; Colly Responsible; +https://github.com/tb0hdan/collyresponsible)",
		ResponseHooks: []func(r *colly.Response){
			func(r *colly.Response) {
				fmt.Printf("URL: %s\nHeaders: %s\nBody length: %d\n", r.Ctx.Get("url"), r.Headers, len(r.Body))
			},
		},
	}
	err := collyresponsible.Crawl(profile)
	if err != nil {
		panic(err)
	}
}
