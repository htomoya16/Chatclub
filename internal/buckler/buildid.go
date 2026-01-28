package buckler

import (
	"context"
	"errors"
	"regexp"
	"time"
)

type buildIDCache struct {
	value     string
	expiresAt time.Time
	ttl       time.Duration
}

func newBuildIDCache(ttl time.Duration) *buildIDCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &buildIDCache{ttl: ttl}
}

func (c *buildIDCache) Get() (string, bool) {
	if c.value == "" {
		return "", false
	}
	if time.Now().After(c.expiresAt) {
		return "", false
	}
	return c.value, true
}

func (c *buildIDCache) Set(value string) {
	c.value = value
	c.expiresAt = time.Now().Add(c.ttl)
}

var buildIDPattern = regexp.MustCompile(`"buildId"\s*:\s*"([^"]+)"`)

func (c *Client) FetchBuildID(ctx context.Context, sid string) (string, error) {
	if v, ok := c.cache.Get(); ok {
		return v, nil
	}

	if sid == "" {
		return "", errors.New("sid required")
	}

	htmlURL := c.cfg.BucklerBaseURL + "/" + c.cfg.Lang + "/profile/" + sid + "/battlelog"
	_, body, err := c.getReturn(ctx, htmlURL)
	if err != nil {
		return "", err
	}
	matches := buildIDPattern.FindSubmatch(body)
	if len(matches) < 2 {
		return "", errors.New("buildId not found")
	}
	buildID := string(matches[1])
	c.cache.Set(buildID)
	return buildID, nil
}
