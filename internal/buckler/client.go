package buckler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client handles Buckler login and data fetches.
type Client struct {
	cfg    Config
	client *http.Client
	cache  *buildIDCache
}

func NewClient(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar: %w", err)
	}

	hc := &http.Client{
		Jar:     jar,
		Timeout: 20 * time.Second,
		// manual redirect handling
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		cfg:    cfg,
		client: hc,
		cache:  newBuildIDCache(cfg.BuildIDTTL),
	}, nil
}

func (c *Client) Config() Config {
	return c.cfg
}

func (c *Client) EnsureLogin(ctx context.Context) error {
	if c.hasBucklerSession() {
		return nil
	}
	return c.Login(ctx)
}

func (c *Client) Login(ctx context.Context) error {
	loginURL, state, err := c.startAuthorize(ctx)
	if err != nil {
		return err
	}

	if err := c.get(ctx, loginURL); err != nil {
		return err
	}

	if err := c.postJSON(ctx, c.authURL("/usernamepassword/challenge"), map[string]string{
		"state": state,
	}); err != nil {
		return err
	}

	csrf := c.cookieValue(c.authBaseURL(), "_csrf")
	payload := map[string]any{
		"client_id":    c.cfg.ClientID,
		"connection":   c.cfg.Connection,
		"password":     c.cfg.Password,
		"popup_options": map[string]any{},
		"protocol":     c.cfg.Protocol,
		"redirect_uri": c.cfg.RedirectURI,
		"response_type": c.cfg.ResponseType,
		"scope":        c.cfg.Scope,
		"show_sing_up": c.cfg.ShowSignUp,
		"sso":          true,
		"state":        state,
		"tenant":       c.cfg.Tenant,
		"ui_locales":   c.cfg.UILocales,
		"username":     c.cfg.Email,
		"_csrf":        csrf,
		"_intstate":    "deprecated",
	}

	resp, body, err := c.postJSONReturn(ctx, c.authURL("/usernamepassword/login"), payload)
	if err != nil {
		return err
	}

	if c.hasBucklerSession() {
		return nil
	}

	if action, fields := findLoginCallbackForm(body); action != "" {
		resp, body, err = c.postFormReturn(ctx, action, fields)
		if err != nil {
			return err
		}
	}

	if loc := getLocation(resp); loc != "" {
		if err := c.followRedirects(ctx, loc); err != nil {
			return err
		}
	}

	if c.hasBucklerSession() {
		return nil
	}

	return errors.New("buckler login failed: no session cookies")
}

func (c *Client) startAuthorize(ctx context.Context) (string, string, error) {
	q := url.Values{}
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURI)
	q.Set("response_type", c.cfg.ResponseType)
	q.Set("scope", c.cfg.Scope)
	q.Set("ui_locales", c.cfg.UILocales)

	authURL := c.authURL("/authorize") + "?" + q.Encode()

	resp, _, err := c.getReturn(ctx, authURL)
	if err != nil {
		return "", "", err
	}
	loc := getLocation(resp)
	if loc == "" {
		// fallback to login URL if authorize did not redirect
		login := c.authURL("/login") + "?" + q.Encode()
		return login, extractState(login), nil
	}

	loginURL := resolveURL(authURL, loc)
	return loginURL, extractState(loginURL), nil
}

func (c *Client) hasBucklerSession() bool {
	u, _ := url.Parse(c.cfg.BucklerBaseURL)
	cookies := c.client.Jar.Cookies(u)
	var hasID, hasRID bool
	for _, ck := range cookies {
		switch ck.Name {
		case "buckler_id":
			hasID = true
		case "buckler_r_id":
			hasRID = true
		}
	}
	return hasID && hasRID
}

func (c *Client) authBaseURL() string {
	return strings.TrimRight(c.cfg.AuthBaseURL, "/")
}

func (c *Client) authURL(path string) string {
	return c.authBaseURL() + path
}

func (c *Client) get(ctx context.Context, rawURL string) error {
	_, _, err := c.getReturn(ctx, rawURL)
	return err
}

func (c *Client) getReturn(ctx context.Context, rawURL string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, err
	}
	c.applyHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

func (c *Client) postJSON(ctx context.Context, rawURL string, payload any) error {
	_, _, err := c.postJSONReturn(ctx, rawURL, payload)
	return err
}

func (c *Client) postJSONReturn(ctx context.Context, rawURL string, payload any) (*http.Response, []byte, error) {
	req, err := newJSONRequest(ctx, rawURL, payload)
	if err != nil {
		return nil, nil, err
	}
	c.applyHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

func (c *Client) postFormReturn(ctx context.Context, rawURL string, fields map[string]string) (*http.Response, []byte, error) {
	req, err := newFormRequest(ctx, rawURL, fields)
	if err != nil {
		return nil, nil, err
	}
	c.applyHeaders(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "ja,en-US;q=0.9,en;q=0.8")
	if req.Header.Get("Origin") == "" {
		if req.URL.Host != "" {
			req.Header.Set("Origin", "https://"+req.URL.Host)
		}
	}
}

func (c *Client) followRedirects(ctx context.Context, startURL string) error {
	current := startURL
	for i := 0; i < 20; i++ {
		resp, _, err := c.getReturn(ctx, current)
		if err != nil {
			return err
		}
		if isRedirect(resp.StatusCode) {
			loc := getLocation(resp)
			if loc == "" {
				return fmt.Errorf("redirect without location: %s", current)
			}
			current = resolveURL(current, loc)
			if c.hasBucklerSession() {
				return nil
			}
			continue
		}
		if c.hasBucklerSession() {
			return nil
		}
		return nil
	}
	return errors.New("redirect chain too long")
}
