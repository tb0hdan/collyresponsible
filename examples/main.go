package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/tb0hdan/collyresponsible"
)

const (
	HTTPTimeout = 10 * time.Second
	MaxRunTime  = 1 * time.Hour
)

var bannedExtensions = []string{
	".jpg", ".jpeg", ".png",
	".flac", ".mp3", ".mp4", ".wav",
	".7z", ".p7zip", ".rar", ".zip",
}

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
		URLTests: []func(targetURL string) bool{
			func(targetURL string) bool {
				parsed, err := url.Parse(targetURL)
				if err != nil {
					return false
				}
				// Skip banned extensions
				for _, banned := range bannedExtensions {
					if strings.HasSuffix(parsed.Path, banned) {
						fmt.Println("Skipping banned extension: ", targetURL)
						return false
					}
				}

				return true
			},
		},
		URLHooks: []func(targetURL string){
			func(targetURL string) {
				fmt.Printf("HookURL: %s\n", targetURL)
			},
		},
	}
	err := collyresponsible.Crawl(profile)
	if err != nil {
		panic(err)
	}
}
