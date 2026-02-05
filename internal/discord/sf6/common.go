package sf6

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/internal/discord/common"
)

func characterImageURL(toolName string) string {
	base := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	if base != "" {
		return base + "/api/sf6/character/" + toolName + ".png"
	}
	return "https://www.streetfighter.com/6/buckler/assets/images/material/character/profile_" + toolName + ".png"
}

func imageExists(ctx context.Context, url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		common.Logf("[sf6][embed] imageExists head: url=%s err=%v", url, err)
		return true
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.streetfighter.com/6/buckler/ja-jp/")
	resp, err := client.Do(req)
	if err != nil {
		common.Logf("[sf6][embed] imageExists head: url=%s err=%v", url, err)
		return true
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		common.Logf("[sf6][embed] imageExists head: url=%s status=%d (not found)", url, resp.StatusCode)
		return false
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		common.Logf("[sf6][embed] imageExists head: url=%s status=%d (ok)", url, resp.StatusCode)
		return true
	}
	if resp.StatusCode == http.StatusForbidden {
		common.Logf("[sf6][embed] imageExists head: url=%s status=%d (forbidden)", url, resp.StatusCode)
		return true
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		common.Logf("[sf6][embed] imageExists head: url=%s status=%d (fallback ok)", url, resp.StatusCode)
		return true
	}
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		common.Logf("[sf6][embed] imageExists get: url=%s err=%v", url, err)
		return true
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://www.streetfighter.com/6/buckler/ja-jp/")
	resp, err = client.Do(req)
	if err != nil {
		common.Logf("[sf6][embed] imageExists get: url=%s err=%v", url, err)
		return true
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		common.Logf("[sf6][embed] imageExists get: url=%s status=%d (not found)", url, resp.StatusCode)
		return false
	}
	common.Logf("[sf6][embed] imageExists get: url=%s status=%d", url, resp.StatusCode)
	return resp.StatusCode >= 200 && resp.StatusCode < 400 || resp.StatusCode == http.StatusForbidden
}

func formatJST(t time.Time) string {
	return common.FormatJST(t)
}

func minInt(a, b int) int {
	return common.MinInt(a, b)
}

func maxInt(a, b int) int {
	return common.MaxInt(a, b)
}
