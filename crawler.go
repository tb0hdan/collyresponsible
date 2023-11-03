package collyresponsible

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func Crawl(ctx context.Context, webSite, userAgent string, options ...[]colly.CollectorOption) (err error) {
	parsed, err := url.Parse(webSite)
	if err != nil {
		return err
	}
	// Get robots.txt
	limiter := NewLimiter(2)
	robots, err := GetRobots(ctx, webSite, userAgent, limiter)
	if err != nil {
		return err
	}
	// Check if the user agent is allowed to visit the website
	if !robots.TestAgent(webSite, userAgent) {
		return fmt.Errorf("User agent is not allowed to visit the website")
	}
	// Sleep after getting robots.txt
	limiter.Sleep()
	//
	visitMap := NewVisitMap()

	collectorOptions := []colly.CollectorOption{
		// colly.Async(),
		colly.UserAgent(userAgent),
	}

	if len(options) > 0 {
		collectorOptions = append(collectorOptions, options[0]...)
	}

	// Instantiate default collector
	c := colly.NewCollector(collectorOptions...)

	// Use empty limit rule for collector
	c.Limit(&colly.LimitRule{})

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

		limiter.Sleep()
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
