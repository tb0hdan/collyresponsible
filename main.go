package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/temoto/robotstxt"
)

type RequestLimiter struct {
	SleepDelay int
	sleepMin   int
	lock       sync.RWMutex
}

func (r *RequestLimiter) Increase() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.SleepDelay++
}

func (r *RequestLimiter) Decrease() {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.SleepDelay > r.sleepMin {
		r.SleepDelay--
	}
}

func (r *RequestLimiter) Sleep() {
	r.lock.RLock()
	defer r.lock.RUnlock()
	time.Sleep(time.Duration(r.SleepDelay) * time.Second)
}

func NewLimiter(sleepDelay int) *RequestLimiter {
	return &RequestLimiter{
		SleepDelay: sleepDelay,
		sleepMin:   sleepDelay,
		lock:       sync.RWMutex{},
	}
}

type VisitMap struct {
	visited map[string]bool
	lock    sync.RWMutex
}

func (v *VisitMap) Add(url string) {
	v.lock.Lock()
	defer v.lock.Unlock()
	v.visited[url] = true
}

func (v *VisitMap) IsVisited(url string) bool {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.visited[url]
}

func NewVisitMap() *VisitMap {
	return &VisitMap{
		visited: make(map[string]bool),
		lock:    sync.RWMutex{},
	}
}

func GetRobots(ctx context.Context, website, userAgent string) (*robotstxt.RobotsData, error) {
	if strings.HasSuffix(website, "/") {
		website = website[:len(website)-1]
	}
	head, err := http.NewRequestWithContext(ctx, http.MethodHead, website+"/robots.txt", nil)
	if err != nil {
		return nil, err
	}
	head.Header.Add("User-Agent", userAgent)
	//
	resp, err := http.DefaultClient.Do(head)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	get, err := http.NewRequestWithContext(ctx, http.MethodGet, website+"/robots.txt", nil)
	if err != nil {
		return nil, err
	}
	get.Header.Add("User-Agent", userAgent)
	//
	getResp, err := http.DefaultClient.Do(get)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	robots, err := robotstxt.FromResponse(getResp)
	if err != nil {
		return nil, err
	}
	return robots, nil
}

func crawl(ctx context.Context, webSite, userAgent string, options ...[]colly.CollectorOption) (err error) {
	parsed, err := url.Parse(webSite)
	if err != nil {
		return err
	}
	// Get robots.txt
	robots, err := GetRobots(ctx, webSite, userAgent)
	if err != nil {
		return err
	}
	// Check if the user agent is allowed to visit the website
	if !robots.TestAgent(webSite, userAgent) {
		return fmt.Errorf("User agent is not allowed to visit the website")
	}

	visitMap := NewVisitMap()
	limiter := NewLimiter(2)

	collectorOptions := []colly.CollectorOption{
		colly.Async(),
		colly.UserAgent(userAgent),
	}

	if len(options) > 0 {
		collectorOptions = append(collectorOptions, options[0]...)
	}

	// Instantiate default collector
	c := colly.NewCollector(collectorOptions...)

	// Limit the number of threads started by colly to two
	c.Limit(&colly.LimitRule{
		Parallelism: 2,
	})

	// Pass down URL from request to response context
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
	})

	// After making a request get "url" from
	// the context of the request
	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("URL: %s\nHeaders: %s\nBody length: %d\n", r.Ctx.Get("url"), r.Headers, len(r.Body))
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// this works
		if strings.ToLower(e.Attr("rel")) == "nofollow" {
			fmt.Println("nofollow: ", link)
			return
		}
		absoluteLink := e.Request.AbsoluteURL(link)
		// Print link
		// Visit link found on page on a new thread
		limiter.Sleep()

		currentHost, err := url.Parse(absoluteLink)
		if err != nil {
			return
		}

		if currentHost.Host != parsed.Host {
			return
		}

		// Check if the user agent is allowed to visit the website
		// absolute links don't work with robotester
		if !robots.TestAgent(link, userAgent) {
			fmt.Println("robots: ", link)
			return
		}

		// Check if the link was already visited
		if visitMap.IsVisited(absoluteLink) {
			return
		}

		fmt.Println("Visiting", absoluteLink)
		c.Visit(absoluteLink)
		visitMap.Add(absoluteLink)
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		if r.StatusCode == http.StatusTooManyRequests {
			limiter.Increase()
		}

		if r.StatusCode == http.StatusOK {
			limiter.Decrease()
		}
	})

	// Start scraping
	c.Visit(webSite)

	// Wait until threads are finished
	runCtx, cancel := context.WithTimeout(context.Background(), 3600*time.Second)
	defer cancel()
	go func(c *colly.Collector) {
		c.Wait()
		// signal the context to cancel itself
		cancel()
	}(c)
	<-runCtx.Done()

	return nil
}

func main() {
	httpCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := crawl(httpCtx, "https://en.wikipedia.org", "Mozilla/5.0 (compatible; Colly Responsible; +https://github.com/tb0hdan/colly-responsible)")
	if err != nil {
		panic(err)
	}
}
