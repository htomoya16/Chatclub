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

// newBuildIDCache は buildId のキャッシュを作る。
func newBuildIDCache(ttl time.Duration) *buildIDCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &buildIDCache{ttl: ttl}
}

// Get は有効なキャッシュがあれば返す。
func (c *buildIDCache) Get() (string, bool) {
	if c.value == "" {
		return "", false
	}
	if time.Now().After(c.expiresAt) {
		return "", false
	}
	return c.value, true
}

// Set は buildId をキャッシュし、期限を更新する。
func (c *buildIDCache) Set(value string) {
	c.value = value
	c.expiresAt = time.Now().Add(c.ttl)
}

var buildIDPattern = regexp.MustCompile(`"buildId"\s*:\s*"([^"]+)"`)

// FetchBuildID は HTML から buildId を抽出する。
// sid が無い場合は Buckler トップから取る。
func (c *Client) FetchBuildID(ctx context.Context, sid string) (string, error) {
	if v, ok := c.cache.Get(); ok {
		return v, nil
	}

	base := c.cfg.BucklerBaseURL + "/" + c.cfg.Lang
	var htmlURL string
	if sid == "" {
		htmlURL = base
	} else {
		htmlURL = base + "/profile/" + sid + "/battlelog"
	}
	resp, body, err := c.getReturn(ctx, htmlURL)
	if err != nil {
		return "", err
	}
	matches := buildIDPattern.FindSubmatch(body)
	if len(matches) < 2 {
		if loc := getLocation(resp); loc != "" {
			nextURL := resolveURL(htmlURL, loc)
			_, body, err = c.getReturn(ctx, nextURL)
			if err != nil {
				return "", err
			}
			matches = buildIDPattern.FindSubmatch(body)
		}
	}
	if len(matches) < 2 {
		return "", errors.New("buildId not found")
	}
	buildID := string(matches[1])
	c.cache.Set(buildID)
	return buildID, nil
}
