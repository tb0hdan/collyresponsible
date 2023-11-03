package main

import (
	"context"
	"time"

	collyresponsible "github.com/tb0hdan/colly-responsible"
)

func main() {
	httpCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := collyresponsible.Crawl(httpCtx, "https://en.wikipedia.org", "Mozilla/5.0 (compatible; Colly Responsible; +https://github.com/tb0hdan/colly-responsible)")
	if err != nil {
		panic(err)
	}
}
