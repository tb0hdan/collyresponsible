package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/tb0hdan/collyresponsible"
)

const (
	HTTPTimeout = 10 * time.Second
	MaxRunTime  = 1 * time.Hour
)

func main() {
	httpCtx, cancel := context.WithTimeout(context.Background(), HTTPTimeout)
	defer cancel()
	profile := &collyresponsible.CrawlerProfile{
		Ctx:        httpCtx,
		Website:    "https://en.wikipedia.org",
		UserAgent:  "Mozilla/5.0 (compatible; Colly Responsible; +https://github.com/tb0hdan/collyresponsible)",
		MaxRuntime: MaxRunTime,
		ResponseHooks: []func(r *colly.Response){
			func(r *colly.Response) {
				fmt.Printf("ResponseURL: %s\nHeaders: %s\nBody length: %d\n", r.Ctx.Get("url"), r.Headers, len(r.Body))
			},
		},
		URLHooks: []func(url string){
			func(url string) {
				fmt.Printf("HookURL: %s\n", url)
			},
		},
	}
	err := collyresponsible.Crawl(profile)
	if err != nil {
		panic(err)
	}
}
