package buckler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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
	return code == http.StatusFound || code == http.StatusSeeOther || code == http.StatusTemporaryRedirect || code == http.StatusMovedPermanently
}

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

func extractState(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get("state")
}

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
