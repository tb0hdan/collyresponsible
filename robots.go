package collyresponsible

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/temoto/robotstxt"
)

func GetRobots(ctx context.Context, website, userAgent string, limiter *RequestLimiter) (*robotstxt.RobotsData, error) {
	if strings.HasSuffix(website, "/") {
		website = website[:len(website)-1]
	}
	head, err := http.NewRequestWithContext(ctx, http.MethodHead, fmt.Sprintf("%s/%s", website, RobotsTxt), nil)
	if err != nil {
		return nil, err
	}
	head.Header.Add(UserAgentHeader, userAgent)
	//
	resp, err := http.DefaultClient.Do(head)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, err
	}
	//
	limiter.Sleep()
	//
	get, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", website, RobotsTxt), nil)
	if err != nil {
		return nil, err
	}
	get.Header.Add(UserAgentHeader, userAgent)
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
