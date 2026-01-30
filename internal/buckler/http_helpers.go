package buckler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// newJSONRequest は JSON で POST するためのリクエストを作る。
func newJSONRequest(ctx context.Context, rawURL string, payload any) (*http.Request, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(payload); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// newFormRequest は form-urlencoded で POST するためのリクエストを作る。
func newFormRequest(ctx context.Context, rawURL string, fields map[string]string) (*http.Request, error) {
	values := url.Values{}
	for k, v := range fields {
		values.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func readBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getLocation(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	return resp.Header.Get("Location")
}

func isRedirect(code int) bool {
	return code == http.StatusFound || code == http.StatusSeeOther || code == http.StatusTemporaryRedirect || code == http.StatusMovedPermanently || code == http.StatusPermanentRedirect
}

// resolveURL は相対URLを基準URLで解決して絶対URLにする。
func resolveURL(baseRaw, refRaw string) string {
	ref, err := url.Parse(refRaw)
	if err != nil {
		return refRaw
	}
	if ref.IsAbs() {
		return ref.String()
	}
	base, err := url.Parse(baseRaw)
	if err != nil {
		return refRaw
	}
	return base.ResolveReference(ref).String()
}

// extractState は URL クエリの state を取り出す。
func extractState(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("state")
}

// isAbsURL は絶対URLかどうかを判定する。
func isAbsURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// cookieValue は CookieJar から指定Cookieを探す。
func (c *Client) cookieValue(rawURL, name string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for _, ck := range c.client.Jar.Cookies(u) {
		if ck.Name == name {
			return ck.Value
		}
	}
	return ""
}

func debugSummary(resp *http.Response) string {
	if resp == nil {
		return "nil response"
	}
	return fmt.Sprintf("status=%d location=%s", resp.StatusCode, resp.Header.Get("Location"))
}

var redirectURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`https?://[^"'\\s>]*loginCallback[^"'\\s>]*`),
	regexp.MustCompile(`https?://[^"'\\s>]*auth\\.cid\\.capcom\\.com/authorize[^"'\\s>]*`),
	regexp.MustCompile(`https?://[^"'\\s>]*streetfighter\\.com/6/buckler/auth/login[^"'\\s>]*`),
}

var redirectRelativePatterns = []*regexp.Regexp{
	regexp.MustCompile(`"/authorize/[^"']+"`),
	regexp.MustCompile(`"/loginCallback[^"']+"`),
}

// findRedirectURLFromHTML は HTML からリダイレクト先っぽいURLを拾う。
// Location ヘッダが無い場合の保険として使う。
func findRedirectURLFromHTML(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	text := string(body)
	for _, re := range redirectURLPatterns {
		if m := re.FindString(text); m != "" {
			return m
		}
	}
	for _, re := range redirectRelativePatterns {
		if m := re.FindString(text); m != "" {
			return strings.Trim(m, "\"")
		}
	}
	return ""
}
