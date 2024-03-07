package collyresponsible

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/temoto/robotstxt"
)

func TestRobotsGroup(robots *robotstxt.RobotsData, url, userAgent string) bool {
	status := robots.TestAgent(url, userAgent)
	if !status {
		return false
	}
	// Test all user agents
	parts := strings.Split(strings.TrimRight(userAgent, ")"), ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// url support
		newPart := strings.TrimPrefix(strings.TrimPrefix(part, "+https://"), "http://")
		// version support
		newPart = strings.TrimSpace(strings.Split(newPart, "/")[0])
		if !robots.TestAgent(url, newPart) {
			return false
		}

	}
	return status
}

func Crawl(profile *CrawlerProfile) (err error) {
	parsed, err := url.Parse(profile.Website)
	if err != nil {
		return err
	}
	// Get robots.txt
	limiter := NewLimiter(2)
	robots, err := GetRobots(profile.Ctx, profile.Website, profile.UserAgent, limiter)
	if err != nil {
		return err
	}
	// Check if the user agent is allowed to visit the website
	if !TestRobotsGroup(robots, profile.Website, profile.UserAgent) {
		return fmt.Errorf("User agent is not allowed to visit the website")
	}

	// Sleep after getting robots.txt
	limiter.Sleep()
	//
	visitMap := NewVisitMap()

	collectorOptions := []colly.CollectorOption{
		// Does not work with Async
		// colly.Async(),
		colly.UserAgent(profile.UserAgent),
		colly.CheckHead(),
	}

	if len(profile.CollyOptions) > 0 {
		collectorOptions = append(collectorOptions, profile.CollyOptions...)
	}

	// Instantiate default collector
	c := colly.NewCollector(collectorOptions...)

	// Use empty limit rule for collector
	if profile.CollyLimits == nil {
		profile.CollyLimits = &colly.LimitRule{DomainGlob: "*"}
	}
	//
	c.Limit(profile.CollyLimits)

	// Pass down URL from request to response context
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
	})

	// After making a request get "url" from
	// the context of the request
	c.OnResponse(func(r *colly.Response) {
		for _, fn := range profile.ResponseHooks {
			fn(r)
		}
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// this works
		if strings.ToLower(e.Attr("rel")) == NoFollow {
			fmt.Printf("%s: %s\n", NoFollow, link)
			return
		}
		absoluteLink := e.Request.AbsoluteURL(link)
		// Print link
		// Visit link found on page on a new thread

		currentHost, err := url.Parse(absoluteLink)
		if err != nil {
			return
		}
		// Always skip already visited links
		if visitMap.IsVisited(absoluteLink) {
			return
		}
		// Pass URL to hooks before checking if the link is allowed
		for _, fn := range profile.URLHooks {
			if absoluteLink != "" {
				fn(absoluteLink)
			}
		}
		//
		if currentHost.Host != parsed.Host {
			return
		}
		// Check if the user agent is allowed to visit the website
		// absolute links don't work with robots.txt tester
		if !robots.TestAgent(link, profile.UserAgent) {
			fmt.Println("robots: ", link)
			return
		}
		// Check if the link is allowed
		for _, fn := range profile.URLTests {
			if !fn(absoluteLink) {
				return
			}
		}
		//
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
	go func() {
		// c.Wait() only makes sense with colly.Async(), so we just start goroutine
		c.Visit(profile.Website)
	}()

	// Wait until threads are finished
	runCtx, cancel := context.WithTimeout(context.Background(), profile.MaxRuntime)
	defer cancel()

	<-runCtx.Done()

	return nil
}
