package collyresponsible

import (
	"context"
	"net/http"
	"strings"

	"github.com/temoto/robotstxt"
)

func GetRobots(ctx context.Context, website, userAgent string, limiter *RequestLimiter) (*robotstxt.RobotsData, error) {
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
	//
	limiter.Sleep()
	//
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
