package buckler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client は Buckler のログインとデータ取得を担当する。
type Client struct {
	cfg    Config
	client *http.Client
	cache  *buildIDCache
}

var errBucklerSessionNotEstablished = errors.New("buckler session not established")

func NewClient(cfg Config) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookie jar: %w", err)
	}

	hc := &http.Client{
		Jar:     jar,
		Timeout: 20 * time.Second,
		// リダイレクトは自前で追いかける
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

// EnsureLogin はログインCookieが無ければログインを実行する。
func (c *Client) EnsureLogin(ctx context.Context) error {
	if c.hasBucklerSession() {
		return nil
	}
	return c.Login(ctx)
}

// Login は Auth0 経由のログインフローを実行する。
func (c *Client) Login(ctx context.Context) error {
	// 1) authorize でログインURLと state を取得
	loginURL, state, err := c.startAuthorize(ctx)
	if err != nil {
		return err
	}

	// 2) ログインページにアクセスして Cookie を受け取る
	if err := c.get(ctx, loginURL); err != nil {
		return err
	}
	fmt.Printf("[buckler] login page ok: %s\n", loginURL)

	// 3) challenge に state を送る
	if err := c.postJSON(ctx, c.authURL("/usernamepassword/challenge"), map[string]string{
		"state": state,
	}); err != nil {
		return err
	}
	fmt.Printf("[buckler] challenge ok\n")

	// 4) username/password を送信（CSRF も Cookie から取得）
	csrf := c.cookieValue(c.authBaseURL(), "_csrf")
	payload := map[string]any{
		"client_id":     c.cfg.ClientID,
		"connection":    c.cfg.Connection,
		"password":      c.cfg.Password,
		"popup_options": map[string]any{},
		"protocol":      c.cfg.Protocol,
		"redirect_uri":  c.cfg.RedirectURI,
		"response_type": c.cfg.ResponseType,
		"scope":         c.cfg.Scope,
		"show_sing_up":  c.cfg.ShowSignUp,
		"sso":           true,
		"state":         state,
		"tenant":        c.cfg.Tenant,
		"ui_locales":    c.cfg.UILocales,
		"username":      c.cfg.Email,
		"_csrf":         csrf,
		"_intstate":     "deprecated",
	}

	resp, body, err := c.postJSONReturn(ctx, c.authURL("/usernamepassword/login"), payload)
	if err != nil {
		return err
	}
	fmt.Printf("[buckler] login submit ok: %s\n", debugSummary(resp))

	if c.hasBucklerSession() {
		return nil
	}

	// 5) HTML内の /login/callback フォームがあれば POST する
	if action, fields := findLoginCallbackForm(body); action != "" {
		actionURL := action
		if !isAbsURL(actionURL) {
			actionURL = resolveURL(c.authBaseURL()+"/", actionURL)
		}
		resp, body, err = c.postFormReturn(ctx, actionURL, fields)
		if err != nil {
			return err
		}
		fmt.Printf("[buckler] login/callback post: %s\n", debugSummary(resp))
	}

	// 6) 以降のリダイレクトを辿って Buckler 側へ到達する
	if loc := getLocation(resp); loc != "" {
		fmt.Printf("[buckler] redirect from login: %s\n", loc)
		if err := c.followRedirects(ctx, loc); err != nil && !errors.Is(err, errBucklerSessionNotEstablished) {
			return err
		}
	} else if next := findRedirectURLFromHTML(body); next != "" {
		fmt.Printf("[buckler] redirect from html: %s\n", next)
		if err := c.followRedirects(ctx, next); err != nil && !errors.Is(err, errBucklerSessionNotEstablished) {
			return err
		}
	}

	if c.hasBucklerSession() {
		return nil
	}

	// 7) Buckler 側の authorize を明示的に叩く（CAPCOMマイページで止まる場合の救済）
	if err := c.authorizeBuckler(ctx); err != nil {
		return err
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
		// authorize でリダイレクトが無い場合は login を使う
		login := c.authURL("/login") + "?" + q.Encode()
		return login, extractState(login), nil
	}

	loginURL := resolveURL(authURL, loc)
	return loginURL, extractState(loginURL), nil
}

// authorizeBuckler は Buckler 用の authorize を叩いて Buckler Cookie 発行まで進める。
func (c *Client) authorizeBuckler(ctx context.Context) error {
	// Buckler が参照する言語Cookieを streetfighter.com に付ける
	c.setBucklerLangCookies()

	// 先に Buckler トップを踏んで言語やサイト側のCookieを整える
	top := strings.TrimRight(c.cfg.BucklerBaseURL, "/") + "/" + c.cfg.Lang + "/"
	resp, body, _ := c.getReturn(ctx, top)
	if resp != nil && isRedirect(resp.StatusCode) {
		loc := getLocation(resp)
		if loc != "" {
			top = resolveURL(top, loc)
			resp, body, _ = c.getReturn(ctx, top)
		}
	}
	if next := findRedirectURLFromHTML(body); next != "" {
		next = resolveURL(top, next)
		fmt.Printf("[buckler] buckler authorize (from html): %s\n", next)
		return c.followRedirects(ctx, next)
	}

	// loginep から Buckler 公式のリダイレクトを取得
	if redir, err := c.fetchLoginEPRedirect(ctx); err == nil && redir != "" {
		fmt.Printf("[buckler] buckler authorize (loginep): %s\n", redir)
		return c.followRedirects(ctx, redir)
	} else if err != nil && c.cfg.Debug {
		fmt.Printf("[buckler][debug] loginep error: %v\n", err)
	}

	// Buckler 側のログイン入口（state 生成）を叩く
	loginEntry := strings.TrimRight(c.cfg.BucklerBaseURL, "/") + "/" + c.cfg.Lang + "/auth/login"
	resp, body, err := c.getReturn(ctx, loginEntry)
	if err == nil && resp != nil {
		fmt.Printf("[buckler] buckler login entry: %s\n", debugSummary(resp))
		if loc := getLocation(resp); loc != "" {
			loc = resolveURL(loginEntry, loc)
			fmt.Printf("[buckler] redirect from buckler login entry: %s\n", loc)
			return c.followRedirects(ctx, loc)
		}
		if next := findRedirectURLFromHTML(body); next != "" {
			next = resolveURL(loginEntry, next)
			fmt.Printf("[buckler] redirect from buckler login entry html: %s\n", next)
			return c.followRedirects(ctx, next)
		}
	}

	if c.cfg.BucklerClientID == "" {
		return errors.New("BUCKLER_CLIENT_ID required")
	}

	q := url.Values{}
	q.Set("client_id", c.cfg.BucklerClientID)
	q.Set("redirect_uri", c.cfg.BucklerRedirect)
	q.Set("response_type", c.cfg.BucklerResponse)
	q.Set("scope", c.cfg.BucklerScope)
	q.Set("audience", c.cfg.BucklerAudience)
	q.Set("ui_locales", c.cfg.UILocales)
	q.Set("state", randomState(24))

	authURL := c.authURL("/authorize") + "?" + q.Encode()
	resp, body, err = c.getReturn(ctx, authURL)
	if err != nil {
		return err
	}
	fmt.Printf("[buckler] buckler authorize: %s\n", debugSummary(resp))
	if loc := getLocation(resp); loc != "" {
		loc = resolveURL(authURL, loc)
		fmt.Printf("[buckler] redirect from buckler authorize: %s\n", loc)
		return c.followRedirects(ctx, loc)
	}
	if next := findRedirectURLFromHTML(body); next != "" {
		next = resolveURL(authURL, next)
		fmt.Printf("[buckler] redirect from buckler authorize html: %s\n", next)
		return c.followRedirects(ctx, next)
	}
	return nil
}

type loginEPResponse struct {
	PageProps struct {
		Redirect string `json:"__N_REDIRECT"`
	} `json:"pageProps"`
}

func (c *Client) fetchLoginEPRedirect(ctx context.Context) (string, error) {
	buildID, err := c.FetchBuildID(ctx, "")
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(c.cfg.BucklerBaseURL, "/")
	q := url.Values{}
	q.Set("redirect_url", "/")
	dataURL := fmt.Sprintf("%s/_next/data/%s/%s/auth/loginep.json?%s",
		base, buildID, c.cfg.Lang, q.Encode(),
	)
	resp, body, err := c.getReturnWithHeaders(ctx, dataURL, map[string]string{
		"x-nextjs-data": "1",
	})
	if err != nil {
		return "", err
	}
	if resp != nil && resp.StatusCode >= 400 {
		return "", fmt.Errorf("loginep status=%d", resp.StatusCode)
	}
	var payload loginEPResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	return payload.PageProps.Redirect, nil
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

func (c *Client) setBucklerLangCookies() {
	u, err := url.Parse(c.cfg.BucklerBaseURL)
	if err != nil || u.Host == "" {
		return
	}
	lang := strings.TrimSpace(c.cfg.Lang)
	if lang == "" {
		lang = "ja-jp"
	}
	base := lang
	if idx := strings.IndexByte(lang, '-'); idx > 0 {
		base = lang[:idx]
	}
	cookies := []*http.Cookie{
		{Name: "pll_language", Value: lang, Path: "/"},
		{Name: "locale", Value: base, Path: "/"},
		{Name: "NEXT_LOCALE", Value: lang, Path: "/"},
	}
	if c.cookieValue(c.cfg.BucklerBaseURL, "buckler_r_id") == "" {
		if rid := randomUUID(); rid != "" {
			cookies = append(cookies, &http.Cookie{Name: "buckler_r_id", Value: rid, Path: "/"})
		}
	}
	c.client.Jar.SetCookies(u, cookies)

	// www なしにも付けておく
	if strings.HasPrefix(u.Host, "www.") {
		alt := *u
		alt.Host = strings.TrimPrefix(u.Host, "www.")
		c.client.Jar.SetCookies(&alt, cookies)
	}
}

func randomState(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if n <= 0 {
		n = 16
	}
	out := make([]byte, n)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		v, err := rand.Int(rand.Reader, max)
		if err != nil {
			out[i] = charset[0]
			continue
		}
		out[i] = charset[v.Int64()]
	}
	return string(out)
}

func randomUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	// UUID v4
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16],
	)
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
	c.debugRequest(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	c.debugResponse(req, resp)
	defer resp.Body.Close()
	body, err := readBody(resp)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}

func (c *Client) getReturnWithHeaders(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, err
	}
	c.applyHeaders(req)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	c.debugRequest(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	c.debugResponse(req, resp)
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
	c.debugRequest(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	c.debugResponse(req, resp)
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
	c.debugRequest(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	c.debugResponse(req, resp)
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
	req.Header.Set("Accept-Language", c.acceptLanguage())
	if req.Header.Get("Referer") == "" && isBucklerHost(req.URL.Host) {
		if !strings.Contains(req.URL.Path, "/auth/login") {
			req.Header.Set("Referer", c.bucklerReferer())
		}
	}
	if isBucklerHost(req.URL.Host) && useNavHeaders(req.URL.Path) {
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("Upgrade-Insecure-Requests", "1")
		req.Header.Set("Sec-Fetch-Dest", "document")
		req.Header.Set("Sec-Fetch-Mode", "navigate")
		if strings.Contains(req.URL.Path, "/auth/login") {
			req.Header.Set("Sec-Fetch-Site", "cross-site")
		} else {
			req.Header.Set("Sec-Fetch-Site", "same-origin")
		}
		req.Header.Set("Sec-Fetch-User", "?1")
	}
	if req.Method != http.MethodGet && req.Header.Get("Origin") == "" {
		if req.URL.Host != "" {
			req.Header.Set("Origin", "https://"+req.URL.Host)
		}
	}
}

func isBucklerHost(host string) bool {
	return strings.Contains(host, "streetfighter.com")
}

func useNavHeaders(path string) bool {
	return strings.HasPrefix(path, "/6/buckler/") && !strings.Contains(path, "/_next/")
}

func (c *Client) bucklerReferer() string {
	base := strings.TrimRight(c.cfg.BucklerBaseURL, "/")
	lang := strings.TrimSpace(c.cfg.Lang)
	if lang == "" {
		lang = "ja-jp"
	}
	return base + "/" + lang + "/"
}

func (c *Client) debugRequest(req *http.Request) {
	if !c.cfg.Debug || req == nil || req.URL == nil {
		return
	}
	host := req.URL.Host
	if !isBucklerHost(host) {
		return
	}
	fmt.Printf("[buckler][debug] req %s %s\n", req.Method, req.URL.String())
	fmt.Printf("[buckler][debug] headers: Accept=%s Accept-Language=%s Referer=%s Origin=%s\n",
		req.Header.Get("Accept"),
		req.Header.Get("Accept-Language"),
		req.Header.Get("Referer"),
		req.Header.Get("Origin"),
	)
	fmt.Printf("[buckler][debug] cookies: %s\n", summarizeCookies(c.client.Jar.Cookies(req.URL)))
}

func (c *Client) debugResponse(req *http.Request, resp *http.Response) {
	if !c.cfg.Debug || req == nil || req.URL == nil || resp == nil {
		return
	}
	if !isBucklerHost(req.URL.Host) {
		return
	}
	fmt.Printf("[buckler][debug] resp %d location=%s\n", resp.StatusCode, resp.Header.Get("Location"))
	if set := resp.Header.Values("Set-Cookie"); len(set) > 0 {
		fmt.Printf("[buckler][debug] set-cookie: %s\n", summarizeSetCookies(set))
	}
}

func summarizeCookies(cookies []*http.Cookie) string {
	if len(cookies) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(cookies))
	for _, ck := range cookies {
		parts = append(parts, fmt.Sprintf("%s(len=%d)", ck.Name, len(ck.Value)))
	}
	return strings.Join(parts, " ")
}

func summarizeSetCookies(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(values))
	for _, v := range values {
		name := v
		if idx := strings.IndexByte(v, '='); idx > 0 {
			name = v[:idx]
		}
		parts = append(parts, name)
	}
	return strings.Join(parts, " ")
}

func (c *Client) acceptLanguage() string {
	lang := strings.TrimSpace(c.cfg.Lang)
	if lang == "" {
		return "ja,en-US;q=0.9,en;q=0.8"
	}
	base := lang
	if idx := strings.IndexByte(lang, '-'); idx > 0 {
		base = lang[:idx]
	}
	return fmt.Sprintf("%s,en-US;q=0.9,en;q=0.8", base)
}

// followRedirects はリダイレクトを辿って Buckler Cookie 発行まで進める。
func (c *Client) followRedirects(ctx context.Context, startURL string) error {
	current := startURL
	base := c.authBaseURL()
	for i := 0; i < 20; i++ {
		current = c.rewriteBucklerLoginURL(current)
		if !isAbsURL(current) {
			current = resolveURL(base, current)
		}
		resp, body, err := c.getReturn(ctx, current)
		if err != nil {
			return err
		}
		fmt.Printf("[buckler] follow %d: %s\n", i+1, debugSummary(resp))
		base = current
		if isRedirect(resp.StatusCode) {
			loc := getLocation(resp)
			if loc == "" {
				return fmt.Errorf("redirect without location: %s", current)
			}
			current = resolveURL(base, loc)
			continue
		}
		if next := findRedirectURLFromHTML(body); next != "" {
			current = resolveURL(base, next)
			continue
		}
		if c.hasBucklerSession() {
			return nil
		}
		return errBucklerSessionNotEstablished
	}
	return errors.New("redirect chain too long")
}

func (c *Client) rewriteBucklerLoginURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || !isBucklerHost(u.Host) {
		return rawURL
	}
	if u.Path == "/6/buckler/auth/login" {
		lang := strings.TrimSpace(c.cfg.Lang)
		if lang == "" {
			lang = "ja-jp"
		}
		u.Path = "/6/buckler/" + lang + "/auth/login"
		return u.String()
	}
	return rawURL
}
